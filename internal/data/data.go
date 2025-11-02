package data

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

// CmdData the data layer command (Repo implementation)
var CmdData = &cobra.Command{
	Use:   "data",
	Short: "Generate the go-kratos data layer implementations",
	Long:  "Generate the go-kratos data layer implementations (Repo interface implementation). Example: kratos proto data api/xxx.proto --target-dir=internal/data --domain-pkg=internal/domain --db-pkg=gorm.io/gorm",
	Run:   run,
}

var (
	targetDir string // 生成目标目录
)

// 初始化命令行参数
func init() {
	CmdData.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/data", "generate target directory")
}

// 核心执行逻辑
func run(_ *cobra.Command, args []string) {
	// 检查 proto 文件参数
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto data api/xxx.proto")
		return
	}

	// 打开 proto 文件
	reader, err := os.Open(args[0])
	if err != nil {
		log.Fatalf("failed to open proto file: %v", err)
	}
	defer reader.Close()

	// 解析 proto 文件（复用 server/biz 的解析逻辑）
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatalf("failed to parse proto file: %v", err)
	}

	// 提取 proto 关键信息（服务 + 方法，补充 PB 包路径）
	dataData := &DataData{}
	proto.Walk(definition,
		// 提取服务和方法（Repo 方法与 biz 层 UseCase 一一对应）
		proto.WithService(func(s *proto.Service) {
			dataData.Service = nameToUpperCamelCase(s.Name)          // 服务名
			dataData.UseCasePackage = convertToBizPackage(targetDir) // 领域层 UseCase 包路径
			// 遍历 RPC 方法，生成 Repo 对应的实现方法
			for _, e := range s.Elements {
				rpc, ok := e.(*proto.RPC)
				if !ok {
					continue
				}

				comment := ""
				if rpc.Comment != nil {
					comment = rpc.Comment.Message()
				}
				// 添加方法信息
				dataData.Methods = append(dataData.Methods, &DataMethod{
					ServiceName: nameToUpperCamelCase(s.Name),
					MethodName:  nameToUpperCamelCase(rpc.Name),
					ParamType:   "*" + cleanNameRequest(cleanTypeName(rpc.RequestType)),
					ParamName:   cleanNameRequestAndToLowerCamelCase(cleanTypeName(rpc.RequestType)),
					ReturnType:  "*" + cleanNameReply(cleanTypeName(rpc.ReturnsType)),
					Comment:     comment,
				})
			}
		}),
	)

	// 检查并创建目标目录
	if _, err = os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exist, creating...\n", targetDir)
		if err = os.MkdirAll(targetDir, 0o755); err != nil {
			log.Fatalf("failed to create target directory: %v", err)
		}
	}

	// 加载并解析 data 层模板
	tpl, err := template.New("dataTemplate").Parse(dataTemplate)
	if err != nil {
		log.Fatalf("failed to parse data template: %v", err)
	}

	// 生成每个服务的 Repo 实现文件
	// 生成文件名：小写服务名 + _repo.go（kratos data 层命名规范）
	filename := strings.ToLower(dataData.Service) + ".go"
	targetPath := filepath.Join(targetDir, filename)

	// 跳过已存在的文件
	if _, err = os.Stat(targetPath); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "data file already exists: %s\n", targetPath)
		return
	}

	// 创建文件并渲染模板
	file, err := os.Create(targetPath)
	if err != nil {
		log.Fatalf("failed to create data file: %v", err)
	}
	defer file.Close()

	if err := tpl.Execute(file, dataData); err != nil {
		log.Fatalf("failed to render data template: %v", err)
	}

	fmt.Printf("generated data file: %s\n", targetPath)
}

// ------------------------------
// 数据结构：适配 data 层模板变量
// ------------------------------
type DataData struct {
	Service        string        // 服务名
	UseCasePackage string        // 领域层 UseCase 包路径
	Methods        []*DataMethod // Repo 方法列表
}

type DataMethod struct {
	ServiceName string // 服务名
	MethodName  string // 方法名
	ParamType   string // 参数类型
	ParamName   string // 参数名
	ReturnType  string // 返回类型
	Comment     string // 注释
}

// serviceName 服务名/方法名转大驼峰
func nameToUpperCamelCase(name string) string {
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

func cleanNameRequestAndToLowerCamelCase(name string) string {
	return toLowerCamelCase(strings.TrimSuffix(name, "Request"))
}

func cleanNameRequest(name string) string {
	return strings.TrimSuffix(name, "Request")
}

func cleanNameReply(name string) string {
	return strings.TrimSuffix(name, "Reply")
}

// toLowerCamelCase 下划线转小驼峰（如 user_create → userCreate）
func toLowerCamelCase(s string) string {
	if len(s) > 0 {
		s = strings.ToLower(s[:1]) + s[1:]
	}
	return s
}

func convertToBizPackage(dataPath string) string {
	// 步骤1：去除前缀 .\ 或 ./
	trimmed := strings.TrimPrefix(dataPath, `.\`)
	trimmed = strings.TrimPrefix(trimmed, `./`)

	// 步骤2：将路径中的 data 替换为 biz（支持多级路径）
	withBiz := strings.ReplaceAll(trimmed, "data", "biz")

	// 步骤3：统一路径分隔符为 /（Go package 路径必须用 /）
	return strings.ReplaceAll(withBiz, `\`, "/")
}
