package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// ServiceInfo contains information about a service parsed from proto
type ServiceInfo struct {
	Name        string
	Package     string
	GoPackage   string
	Methods     []MethodInfo
	Messages    []MessageInfo
	ImportPaths []string
}

// MethodInfo contains information about a service method
type MethodInfo struct {
	Name       string
	InputType  string
	OutputType string
	InputName  string
	OutputName string
}

// MessageInfo contains information about a proto message
type MessageInfo struct {
	Name   string
	Fields []FieldInfo
}

// FieldInfo contains information about a message field
type FieldInfo struct {
	Name     string
	Type     string
	Number   int32
	TypeName string // For message/enum types
}

// ParseProtoFile parses a proto file and returns service information
func ParseProtoFile(protoFile string) (*ServiceInfo, error) {
	absPath, err := filepath.Abs(protoFile)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// 解析 proto 文件为 FileDescriptorProto
	fd := &descriptorpb.FileDescriptorProto{}
	if err := proto.Unmarshal(content, fd); err != nil {
		return nil, err
	}

	// 构建代码生成请求（pluginReq）
	pluginReq := &pluginpb.CodeGeneratorRequest{
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fd},
		FileToGenerate: []string{fd.GetName()},
		Parameter:      proto.String("paths=source_relative"),
	}

	// 将 pluginReq 序列化为二进制数据
	reqData, err := proto.Marshal(pluginReq)
	if err != nil {
		return nil, fmt.Errorf("marshal plugin request failed: %v", err)
	}

	// 创建管道：r 是读取端，w 是写入端
	r, w, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create pipe failed: %v", err)
	}
	defer r.Close()

	// 启动 goroutine 向管道写入请求数据（模拟标准输入）
	go func() {
		defer w.Close() // 写完后关闭写入端，避免读取端阻塞
		if _, err := w.Write(reqData); err != nil {
			fmt.Printf("write to pipe failed: %v\n", err)
		}
	}()

	var serviceInfo *ServiceInfo
	var runErr error

	// 保存原始标准输入，用完后恢复（避免影响其他逻辑）
	originalStdin := os.Stdin
	os.Stdin = r // 将标准输入重定向到管道的读取端

	// 使用 protogen.Options.Run 处理请求（此时会从管道读取数据）
	opts := protogen.Options{
		ParamFunc: func(name, value string) error {
			return nil // 接受所有参数
		},
	}

	// 声明一个错误变量，用于在回调函数中捕获错误
	var parseErr error

	// 直接调用 opts.Run，不赋值给任何变量（因为它没有返回值）
	opts.Run(func(gen *protogen.Plugin) error {
		defer func() {
			// 捕获回调中可能的 panic（如有错误处理逻辑，可在此处转换为 error）
			if r := recover(); r != nil {
				parseErr = fmt.Errorf("parse proto panic: %v", r)
			}
		}()

		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}

			if serviceInfo == nil {
				// 处理 Package（*string -> string）
				packageName := ""
				if file.Proto.Package != nil {
					packageName = *file.Proto.Package
				}

				// 处理 GoPackage（*string -> string）
				goPackageName := ""
				if file.Proto.Options != nil && file.Proto.Options.GoPackage != nil {
					goPackageName = *file.Proto.Options.GoPackage
				}

				serviceInfo = &ServiceInfo{
					Package:   packageName,   // 使用处理后的字符串
					GoPackage: goPackageName, // 使用处理后的字符串
				}
			}

			// 解析服务和方法
			for _, service := range file.Services {
				// 服务名称：service.Desc.Name() 返回 protoreflect.Name，转换为 string
				serviceInfo.Name = string(service.Desc.Name())

				for _, method := range service.Methods {
					// 方法名称：method.Desc.Name() 返回 protoreflect.Name，转换为 string
					methodName := string(method.Desc.Name())

					// 输入类型：先通过 Input() 获取消息描述符，再通过 FullName() 获取完整名称
					inputMsgDesc := method.Desc.Input()          // 返回 protoreflect.MessageDescriptor
					inputType := string(inputMsgDesc.FullName()) // 完整类型名（如 ".package.ClassRequest"）

					// 输出类型：同理
					outputMsgDesc := method.Desc.Output()
					outputType := string(outputMsgDesc.FullName())

					serviceInfo.Methods = append(serviceInfo.Methods, MethodInfo{
						Name:       methodName,
						InputType:  inputType,
						OutputType: outputType,
						InputName:  method.Input.GoIdent.GoName,  // 正确：GoIdent.GoName 是 string
						OutputName: method.Output.GoIdent.GoName, // 正确
					})
				}
			}

			// 解析消息和字段
			for _, msg := range file.Messages {
				// 消息名称：msg.Desc.Name() 返回 string
				msgName := msg.Desc.Name()
				messageInfo := MessageInfo{Name: msgName}

				for _, field := range msg.Fields {
					// 字段名称：field.Desc.Name() 返回 string
					fieldName := field.Desc.Name()

					// 字段类型名称：field.Desc.TypeName() 返回 string（可能为空，非指针）
					typeName := field.Desc.TypeName()

					fieldInfo := FieldInfo{
						Name:     fieldName,
						Number:   field.Desc.Number(), // 字段编号：Number() 方法返回 int32
						Type:     getFieldType(field), // 需要同步修改 getFieldType 函数
						TypeName: typeName,
					}
					messageInfo.Fields = append(messageInfo.Fields, fieldInfo)
				}
				serviceInfo.Messages = append(serviceInfo.Messages, messageInfo)
			}

			// 解析导入路径
			imports := file.Desc.Imports() // 获取所有导入的文件描述符列表
			for _, impDesc := range imports {
				// impDesc.Path() 返回 protoreflect.FilePath，转换为 string 即导入路径
				importPath := string(impDesc.Path())
				serviceInfo.ImportPaths = append(serviceInfo.ImportPaths, importPath)
			}
		}

		// 回调函数可以返回 error，但 Run 方法会忽略它，因此通过闭包变量传递
		return nil
	})

	// 检查解析过程中是否有错误
	if parseErr != nil {
		return nil, parseErr
	}

	if serviceInfo == nil {
		return nil, fmt.Errorf("no service info parsed from proto")
	}

	// 恢复原始标准输入
	os.Stdin = originalStdin

	// 检查错误
	if runErr != nil {
		return nil, runErr
	}
	if serviceInfo == nil {
		return nil, fmt.Errorf("no service info parsed from proto")
	}

	return serviceInfo, nil
}

// Helper function to get field type
func getFieldType(field *protogen.Field) string {
	switch field.Proto.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return "group"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "message"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return "enum"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	default:
		return "unknown"
	}
}
