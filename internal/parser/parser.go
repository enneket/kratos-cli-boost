package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
)

// 数据结构定义（不变）
type ServiceInfo struct {
	Name        string        // 服务名
	Package     string        // proto包名
	GoPackage   string        // Go包路径
	Methods     []MethodInfo  // 方法列表
	Messages    []MessageInfo // 消息列表
	ImportPaths []string      // 导入路径
}

type MethodInfo struct {
	Name       string // 方法名
	InputType  string // 输入类型完整名
	OutputType string // 输出类型完整名
	InputName  string // 输入类型短名
	OutputName string // 输出类型短名
}

type MessageInfo struct {
	Name   string      // 消息名
	Fields []FieldInfo // 字段列表
}

type FieldInfo struct {
	Name     string // 字段名
	Type     string // 字段类型
	Number   int32  // 字段编号
	TypeName string // 自定义类型名
}

// ParseProtoFile 适配旧版本 emicklei/proto（v0.x）
func ParseProtoFile(protoFile string) (*ServiceInfo, error) {
	// 1. 读取文件
	absPath, err := filepath.Abs(protoFile)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %v", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 2. 解析AST（旧版本返回 *proto.Proto）
	parser := proto.NewParser(strings.NewReader(string(content)))
	protoAST, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("解析proto语法失败: %v", err)
	}

	// 3. 初始化结果和访问者
	serviceInfo := &ServiceInfo{}
	visitor := &protoVisitor{serviceInfo: serviceInfo}

	// 4. 遍历所有元素（含包声明）
	for _, elem := range protoAST.Elements {
		// 先处理包声明（单独提取，不依赖访问者）
		if pkg, ok := elem.(*proto.Package); ok {
			serviceInfo.Package = pkg.Name // 提取包名
			fmt.Printf("[调试] 提取包名: %s\n", serviceInfo.Package)
		}

		// 其他元素通过访问者处理
		if visitee, ok := elem.(proto.Visitee); ok {
			visitee.Accept(visitor)
		}
	}

	return serviceInfo, nil
}

// protoVisitor 适配旧版本的访问者
type protoVisitor struct {
	serviceInfo *ServiceInfo
	currentMsg  *MessageInfo
}

// 处理导入（旧版本 Import 结构）
func (v *protoVisitor) VisitImport(i *proto.Import) {
	path := strings.Trim(i.Filename, "\"")
	v.serviceInfo.ImportPaths = append(v.serviceInfo.ImportPaths, path)
}

// 处理选项（旧版本 Option 结构）
// 处理选项（如 go_package）
// 处理选项（如 go_package）
func (v *protoVisitor) VisitOption(o *proto.Option) {
	if o.Name == "go_package" {
		// 直接访问 Source 字段获取原始字符串（含引号）
		constantSource := o.Constant.Source
		// 去除引号，得到实际的包路径
		v.serviceInfo.GoPackage = strings.Trim(constantSource, "\"")
		fmt.Printf("[调试] 提取 go_package: %s\n", v.serviceInfo.GoPackage)
	}
}

// 处理服务
func (v *protoVisitor) VisitService(s *proto.Service) {
	v.serviceInfo.Name = s.Name
}

// 处理RPC方法
func (v *protoVisitor) VisitRPC(r *proto.RPC) {
	inputType := strings.TrimPrefix(r.RequestType, ".")
	outputType := strings.TrimPrefix(r.ReturnsType, ".")
	v.serviceInfo.Methods = append(v.serviceInfo.Methods, MethodInfo{
		Name:       r.Name,
		InputType:  inputType,
		OutputType: outputType,
		InputName:  getShortTypeName(inputType),
		OutputName: getShortTypeName(outputType),
	})
}

// 处理消息（修复指针问题）
func (v *protoVisitor) VisitMessage(m *proto.Message) {
	v.serviceInfo.Messages = append(v.serviceInfo.Messages, MessageInfo{Name: m.Name})
	v.currentMsg = &v.serviceInfo.Messages[len(v.serviceInfo.Messages)-1]
}

// 处理普通字段（修复方法名和参数类型）
func (v *protoVisitor) VisitNormalField(f *proto.NormalField) {
	if v.currentMsg == nil {
		return // 不在消息内的字段，忽略
	}

	// 字段名
	fieldName := f.Name

	// 字段类型（处理 repeated 修饰符）
	fieldType := f.Type
	if f.Repeated {
		fieldType = "[]" + fieldType // 如 repeated string → []string
	}

	// 字段编号（转换为 int32）
	fieldNumber := int32(f.Sequence)

	// 自定义类型名（仅非基础类型需要）
	var typeName string
	if !isBasicType(fieldType) {
		typeName = fieldType
	}

	// 添加到当前消息的字段列表
	v.currentMsg.Fields = append(v.currentMsg.Fields, FieldInfo{
		Name:     fieldName,
		Type:     fieldType,
		Number:   fieldNumber,
		TypeName: typeName,
	})
}

// 必须实现的空方法（旧版本接口要求）
// 必须实现的所有 Visitor 方法（含空实现）
func (v *protoVisitor) VisitSyntax(s *proto.Syntax)         {}
func (v *protoVisitor) VisitComment(c *proto.Comment)       {}
func (v *protoVisitor) VisitPackage(p *proto.Package)       {}
func (v *protoVisitor) VisitOneof(o *proto.Oneof)           {}
func (v *protoVisitor) VisitEnum(e *proto.Enum)             {}
func (v *protoVisitor) VisitEnumField(f *proto.EnumField)   {}
func (v *protoVisitor) VisitMapField(f *proto.MapField)     {}
func (v *protoVisitor) VisitGroup(g *proto.Group)           {}
func (v *protoVisitor) VisitReserved(r *proto.Reserved)     {}
func (v *protoVisitor) VisitExtensions(e *proto.Extensions) {} // 新增：解决当前错误
func (v *protoVisitor) VisitOneofField(f *proto.OneOfField) {}

// 辅助函数（不变）
func getShortTypeName(fullType string) string {
	parts := strings.Split(fullType, ".")
	return parts[len(parts)-1]
}

func isBasicType(t string) bool {
	basic := map[string]bool{
		"bool": true, "int32": true, "int64": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "string": true, "bytes": true,
		"[]bool": true, "[]int32": true, "[]int64": true, "[]uint32": true, "[]uint64": true,
		"[]float32": true, "[]float64": true, "[]string": true, "[]bytes": true,
	}
	return basic[t]
}
