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

	// 提取 proto 关键信息

	bizData := new(BizData)

	proto.Walk(definition,
		// 提取服务和方法信息
		proto.WithService(func(s *proto.Service) {

			bizData.ServiceName = nameToUpperCamelCase(s.Name)

			// 遍历服务下的 RPC 方法
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
				bizData.Methods = append(bizData.Methods, &BizMethod{
					ServiceName: nameToUpperCamelCase(s.Name),
					MethodName:  nameToUpperCamelCase(rpc.Name),
					ParamType:   "*" + cleanNameRequest(cleanTypeName(rpc.RequestType)),
					ParamName:   cleanNameRequestAndToLowerCamelCase(cleanTypeName(rpc.RequestType)),
					ReturnType:  "*" + cleanNameReply(cleanTypeName(rpc.ReturnsType)),
					Comment:     comment,
				})
				bizData.Entities = append(bizData.Entities, &BizEntity{
					Name: cleanNameRequest(cleanTypeName(rpc.RequestType)),
				})
			}
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
	tpl, err := template.New("bizTemplate").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(bizTemplate)
	if err != nil {
		log.Fatalf("failed to parse biz template: %v", err)
	}

	// 生成 biz 代码文件
	// 生成文件名：小写服务名 + _ucase.go（kratos  biz 层命名规范）
	filename := strings.ToLower(bizData.ServiceName) + ".go"
	targetPath := filepath.Join(targetDir, filename)

	// 检查文件是否已存在
	if _, err = os.Stat(targetPath); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "biz file already exists: %s\n", targetPath)
		return
	}

	// 创建文件并渲染模板
	file, err := os.Create(targetPath)
	if err != nil {
		log.Fatalf("failed to create biz file: %v", err)
	}
	defer file.Close()

	if err := tpl.Execute(file, bizData); err != nil {
		log.Fatalf("failed to render biz template: %v", err)
	}

	fmt.Printf("generated biz file: %s\n", targetPath)
}

// ------------------------------
// 数据结构：适配 biz 层模板变量
// ------------------------------
type BizData struct {
	ServiceName string       // 服务名（大驼峰）
	Methods     []*BizMethod // 方法列表
	Entities    []*BizEntity // 实体列表
}

type BizMethod struct {
	ServiceName string // 服务名
	MethodName  string // 方法名
	ParamType   string // 参数类型
	ParamName   string // 参数名
	ReturnType  string // 返回类型
	Comment     string // 注释
}

type BizEntity struct {
	Name   string // 实体名
	Fields []*BizField
}

type BizField struct {
	FieldName string // 字段名
	FieldType string // 字段类型
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
