package biz

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"text/template"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CmdBiz the biz layer command.
var CmdBiz = &cobra.Command{
	Use:   "biz",
	Short: "Generate the go-kratos biz layer implementations",
	Long:  "Generate the go-kratos biz layer implementations (UseCase + validation). Example: kratos proto biz api/xxx.proto --target-dir=internal/biz --domain-pkg=internal/domain",
	Run:   run,
}

var (
	targetDir string // 生成目标目录
)

// 初始化命令行参数
func init() {
	CmdBiz.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/biz", "generate target directory")
}

// 核心执行逻辑
func run(_ *cobra.Command, args []string) {
	// 检查 proto 文件参数
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto biz api/xxx.proto")
		return
	}

	// 打开 proto 文件
	reader, err := os.Open(args[0])
	if err != nil {
		log.Fatalf("failed to open proto file: %v", err)
	}
	defer reader.Close()

	// 解析 proto 文件
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatalf("failed to parse proto file: %v", err)
	}

	// 提取 proto 关键信息（go_package + 服务 + 方法）
	var (
		protoPkg string     // proto 生成的 PB 包路径（go_package）
		res      []*BizData // 生成模板所需的数据
	)

	proto.Walk(definition,
		// 提取 go_package（PB 包路径）
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				// 处理 "path;alias" 格式，取路径部分
				pkgParts := strings.Split(o.Constant.Source, ";")
				if len(pkgParts) > 0 {
					protoPkg = strings.Trim(pkgParts[0], "\"")
				}
			}
		}),
		// 提取服务和方法信息
		proto.WithService(func(s *proto.Service) {
			bizData := &BizData{
				Service:      serviceName(s.Name), // 服务名（大驼峰）
				ProtoPackage: protoPkg,            // PB 包路径
			}

			// 遍历服务下的 RPC 方法
			for _, e := range s.Elements {
				rpc, ok := e.(*proto.RPC)
				if !ok {
					continue
				}

				// 添加方法信息
				bizData.Methods = append(bizData.Methods, &BizMethod{
					Service: serviceName(s.Name),
					Name:    serviceName(rpc.Name),
					Request: cleanTypeName(rpc.RequestType), // 清理请求类型名（去除包前缀）
					Reply:   cleanTypeName(rpc.ReturnsType), // 清理响应类型名
					Type:    getMethodType(rpc.StreamsRequest, rpc.StreamsReturns),
				})
			}

			res = append(res, bizData)
		}),
	)

	// 检查目标目录是否存在
	if _, err = os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exist, creating...\n", targetDir)
		if err = os.MkdirAll(targetDir, 0o755); err != nil {
			log.Fatalf("failed to create target directory: %v", err)
		}
	}

	// 加载并解析 biz 层模板
	tpl, err := template.New("bizTemplate").Parse(bizTemplate)
	if err != nil {
		log.Fatalf("failed to parse biz template: %v", err)
	}

	// 生成每个服务的 biz 代码文件
	for _, data := range res {
		// 生成文件名：小写服务名 + _ucase.go（kratos  biz 层命名规范）
		filename := strings.ToLower(data.Service) + "_ucase.go"
		targetPath := filepath.Join(targetDir, filename)

		// 检查文件是否已存在
		if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "biz file already exists: %s\n", targetPath)
			continue
		}

		// 创建文件并渲染模板
		file, err := os.Create(targetPath)
		if err != nil {
			log.Fatalf("failed to create biz file: %v", err)
		}
		defer file.Close()

		if err := tpl.Execute(file, data); err != nil {
			log.Fatalf("failed to render biz template: %v", err)
		}

		fmt.Printf("generated biz file: %s\n", targetPath)
	}
}

// ------------------------------
// 数据结构：适配 biz 层模板变量
// ------------------------------
type BizData struct {
	Service      string       // 服务名（大驼峰）
	ProtoPackage string       // PB 包路径（如 github.com/xxx/api/user/v1）
	Methods      []*BizMethod // 方法列表
}

type BizMethod struct {
	Service string     // 服务名
	Name    string     // 方法名（大驼峰）
	Request string     // 请求类型名（纯类型，无包前缀）
	Reply   string     // 响应类型名（纯类型，无包前缀）
	Type    MethodType // 方法类型（普通/流式）
}

// ------------------------------
// 枚举：方法类型（和 service 层保持一致）
// ------------------------------
type MethodType int

const (
	unaryType          MethodType = iota + 1 // 1: 普通方法（非流式）
	twoWayStreamsType                        // 2: 双向流
	requestStreamsType                       // 3: 请求流
	returnsStreamsType                       // 4: 响应流
)

// ------------------------------
// 工具函数（复用原有逻辑，适配 biz 需求）
// ------------------------------
// getMethodType 判断方法类型（流式/普通）
func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}

// serviceName 服务名/方法名转大驼峰
func serviceName(name string) string {
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

// toUpperCamelCase 下划线转大驼峰（如 user_create → UserCreate）
func toUpperCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}

// cleanTypeName 清理类型名（去除前缀点号和包路径，仅保留纯类型名）
// 示例：".user.v1.CreateUserRequest" → "CreateUserRequest"
func cleanTypeName(name string) string {
	name = strings.TrimPrefix(name, ".")
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}
