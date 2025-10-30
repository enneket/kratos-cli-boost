package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"kratos_cli_boost/internal/parser"
)

// GenerateBiz generates biz layer code from proto file
func GenerateBiz(protoFile, targetDir string) error {
	// Parse the proto file
	serviceInfo, err := parser.ParseProtoFile(protoFile)
	if err != nil {
		return fmt.Errorf("failed to parse proto file: %v", err)
	}

	// Create output directory
	outputDir := getOutputPath(targetDir, "biz", serviceInfo)
	if err := createDirectoryIfNotExists(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate service interface
	if err := generateBizService(outputDir, serviceInfo); err != nil {
		return fmt.Errorf("failed to generate biz service: %v", err)
	}

	// Generate service implementation
	if err := generateBizServiceImpl(outputDir, serviceInfo); err != nil {
		return fmt.Errorf("failed to generate biz service implementation: %v", err)
	}

	// Generate domain models
	if err := generateBizModels(outputDir, serviceInfo); err != nil {
		return fmt.Errorf("failed to generate biz models: %v", err)
	}

	return nil
}

// generateBizService generates the biz service interface
func generateBizService(outputDir string, serviceInfo *parser.ServiceInfo) error {
	packageName := getPackageName(serviceInfo.GoPackage)
	fileName := fmt.Sprintf("%s_service.go", strings.ToLower(serviceInfo.Name))
	filePath := filepath.Join(outputDir, fileName)

	tplContent := `package {{.PackageName}}

import (
	"context"
	{{range .ImportPaths}}
	"{{.}}"
	{{end}}
)

// {{.ServiceName}}Service is the biz service interface for {{.ServiceName}}
type {{.ServiceName}}Service interface {
	{{range .Methods}}
	{{.Name}}(ctx context.Context, req *{{.InputName}}) (*{{.OutputName}}, error)
	{{end}}
}
`

	tpl, err := template.New("bizService").Parse(tplContent)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		PackageName string
		ServiceName string
		Methods     []parser.MethodInfo
		ImportPaths []string
	}{
		PackageName: packageName,
		ServiceName: serviceInfo.Name,
		Methods:     serviceInfo.Methods,
		ImportPaths: serviceInfo.ImportPaths,
	}

	return tpl.Execute(file, data)
}

// generateBizServiceImpl generates the biz service implementation
func generateBizServiceImpl(outputDir string, serviceInfo *parser.ServiceInfo) error {
	packageName := getPackageName(serviceInfo.GoPackage)
	fileName := fmt.Sprintf("%s_service_impl.go", strings.ToLower(serviceInfo.Name))
	filePath := filepath.Join(outputDir, fileName)

	tplContent := `package {{.PackageName}}

import (
	"context"
	{{range .ImportPaths}}
	"{{.}}"
	{{end}}
)

// {{.ServiceName}}ServiceImpl is the implementation of {{.ServiceName}}Service
type {{.ServiceName}}ServiceImpl struct {
	// Add dependencies here, e.g. data access layer
	// repo {{.ServiceName}}Repo
}

// New{{.ServiceName}}Service creates a new {{.ServiceName}}Service
func New{{.ServiceName}}Service(
	// Add dependencies as parameters
) {{.ServiceName}}Service {
	return &{{.ServiceName}}ServiceImpl{
		// Initialize dependencies
	}
}

{{range .Methods}}
// {{.Name}} implements {{$.ServiceName}}Service
func (s *{{$.ServiceName}}ServiceImpl) {{.Name}}(ctx context.Context, req *{{.InputName}}) (*{{.OutputName}}, error) {
	// Implement business logic here
	return &{{.OutputName}}{}, nil
}
{{end}}
`

	tpl, err := template.New("bizServiceImpl").Parse(tplContent)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		PackageName string
		ServiceName string
		Methods     []parser.MethodInfo
		ImportPaths []string
	}{
		PackageName: packageName,
		ServiceName: serviceInfo.Name,
		Methods:     serviceInfo.Methods,
		ImportPaths: serviceInfo.ImportPaths,
	}

	return tpl.Execute(file, data)
}

// generateBizModels generates the domain models
func generateBizModels(outputDir string, serviceInfo *parser.ServiceInfo) error {
	packageName := getPackageName(serviceInfo.GoPackage)
	fileName := "model.go"
	filePath := filepath.Join(outputDir, fileName)

	tplContent := `package {{.PackageName}}

// Domain models generated from proto messages

{{range .Messages}}
// {{.Name}} is the domain model for {{.Name}}
type {{.Name}} struct {
	{{range .Fields}}
	{{.Name}} {{.Type}} {{if .TypeName}}// {{.TypeName}}{{end}}
	{{end}}
}
{{end}}
`

	tpl, err := template.New("bizModels").Parse(tplContent)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		PackageName string
		Messages    []parser.MessageInfo
	}{
		PackageName: packageName,
		Messages:    serviceInfo.Messages,
	}

	return tpl.Execute(file, data)
}
