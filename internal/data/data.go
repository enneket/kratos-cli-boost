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
	domainPkg string // 领域层包路径（依赖 Repo 接口）
	protoPkg  string // PB 包路径（可选，用于实体转换）
	dbPkg     string // 数据库包路径（如 gorm、sqlx）
	cachePkg  string // 缓存包路径（可选，如 redis）
	useLogger bool   // 是否启用 kratos 日志（默认启用）
)

// 初始化命令行参数（对齐 server/biz 命令风格）
func init() {
	CmdData.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/data", "generate target directory")
	CmdData.Flags().StringVarP(&domainPkg, "domain-pkg", "d", "internal/domain", "domain layer package path (for Repo interface)")
	CmdData.Flags().StringVarP(&protoPkg, "proto-pkg", "p", "", "proto PB package path (for entity conversion, optional)")
	CmdData.Flags().StringVarP(&dbPkg, "db-pkg", "b", "gorm.io/gorm", "database package path (e.g. gorm.io/gorm, database/sql)")
	CmdData.Flags().StringVarP(&cachePkg, "cache-pkg", "c", "", "cache package path (e.g. github.com/redis/go-redis/v9, optional)")
	CmdData.Flags().BoolVar(&useLogger, "use-logger", true, "whether to use kratos logger")
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
	var res []*DataData // 模板渲染所需数据
	proto.Walk(definition,
		// 提取 go_package（用于 PB 包路径，可选）
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" && protoPkg == "" {
				pkgParts := strings.Split(o.Constant.Source, ";")
				if len(pkgParts) > 0 {
					protoPkg = strings.Trim(pkgParts[0], "\"")
				}
			}
		}),
		// 提取服务和方法（Repo 方法与 biz 层 UseCase 一一对应）
		proto.WithService(func(s *proto.Service) {
			data := &DataData{
				Service:       serviceName(s.Name), // 服务名（大驼峰）
				DomainPackage: domainPkg,           // 领域层包路径
				ProtoPackage:  protoPkg,            // PB 包路径（可选）
				DbPackage:     dbPkg,               // 数据库包路径
				CachePackage:  cachePkg,            // 缓存包路径（可选）
				UseLogger:     useLogger,           // 是否启用日志
			}

			// 遍历 RPC 方法，生成 Repo 对应的实现方法
			for _, e := range s.Elements {
				rpc, ok := e.(*proto.RPC)
				if !ok {
					continue
				}

				data.Methods = append(data.Methods, &DataMethod{
					RepoName:   serviceName(s.Name) + "Repo", // Repo 接口名（如 UserServiceRepo）
					MethodName: serviceName(rpc.Name),        // 方法名（大驼峰）
					// Repo 方法参数：通常接收 domain 实体，返回 domain 实体 + 错误
					ParamType:  domainPkg + "." + cleanTypeName(rpc.RequestType) + "Entity",
					ReturnType: domainPkg + "." + cleanTypeName(rpc.ReturnsType) + "Entity",
				})
			}

			res = append(res, data)
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
	for _, data := range res {
		// 生成文件名：小写服务名 + _repo.go（kratos data 层命名规范）
		filename := strings.ToLower(data.Service) + "_repo.go"
		targetPath := filepath.Join(targetDir, filename)

		// 跳过已存在的文件
		if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "data file already exists: %s\n", targetPath)
			continue
		}

		// 创建文件并渲染模板
		file, err := os.Create(targetPath)
		if err != nil {
			log.Fatalf("failed to create data file: %v", err)
		}
		defer file.Close()

		if err := tpl.Execute(file, data); err != nil {
			log.Fatalf("failed to render data template: %v", err)
		}

		fmt.Printf("generated data file: %s\n", targetPath)
	}
}

// ------------------------------
// 数据结构：适配 data 层模板变量
// ------------------------------
type DataData struct {
	Service       string        // 服务名（大驼峰，如 UserService）
	DomainPackage string        // 领域层包路径（如 internal/domain）
	ProtoPackage  string        // PB 包路径（如 github.com/xxx/api/user/v1）
	DbPackage     string        // 数据库包路径（如 gorm.io/gorm）
	CachePackage  string        // 缓存包路径（可选）
	UseLogger     bool          // 是否启用 kratos 日志
	Methods       []*DataMethod // Repo 方法列表
}

type DataMethod struct {
	RepoName   string // Repo 接口名（如 UserServiceRepo）
	MethodName string // 方法名（如 CreateUser）
	ParamType  string // 参数类型（domain 实体，如 internal/domain.UserEntity）
	ReturnType string // 返回类型（domain 实体，如 internal/domain.UserEntity）
}

// ------------------------------
// 工具函数：复用 server/biz 命令的逻辑
// ------------------------------
func serviceName(name string) string {
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

func toUpperCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}

func cleanTypeName(name string) string {
	name = strings.TrimPrefix(name, ".")
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}
