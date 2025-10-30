package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"kratos_cli_boost/internal/parser"
)

// GenerateData generates data layer code from proto file
func GenerateData(protoFile, targetDir string) error {
	// Parse the proto file
	serviceInfo, err := parser.ParseProtoFile(protoFile)
	if err != nil {
		return fmt.Errorf("failed to parse proto file: %v", err)
	}

	// Create output directory
	outputDir := getOutputPath(targetDir, "data", serviceInfo)
	if err := createDirectoryIfNotExists(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate repository interface
	if err := generateDataRepo(outputDir, serviceInfo); err != nil {
		return fmt.Errorf("failed to generate data repository: %v", err)
	}

	// Generate repository implementation
	if err := generateDataRepoImpl(outputDir, serviceInfo); err != nil {
		return fmt.Errorf("failed to generate data repository implementation: %v", err)
	}

	// Generate data models
	if err := generateDataModels(outputDir, serviceInfo); err != nil {
		return fmt.Errorf("failed to generate data models: %v", err)
	}

	return nil
}

// generateDataRepo generates the data repository interface
func generateDataRepo(outputDir string, serviceInfo *parser.ServiceInfo) error {
	packageName := getPackageName(serviceInfo.GoPackage)
	fileName := fmt.Sprintf("%s_repo.go", strings.ToLower(serviceInfo.Name))
	filePath := filepath.Join(outputDir, fileName)

	tplContent := `package {{.PackageName}}

import (
	"context"
	{{range .ImportPaths}}
	"{{.}}"
	{{end}}
)

// {{.ServiceName}}Repo is the repository interface for {{.ServiceName}}
type {{.ServiceName}}Repo interface {
	{{range .Methods}}
	{{.Name}}(ctx context.Context, req *{{.InputName}}) (*{{.OutputName}}, error)
	{{end}}
}
`

	tpl, err := template.New("dataRepo").Parse(tplContent)
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

// generateDataRepoImpl generates the data repository implementation
func generateDataRepoImpl(outputDir string, serviceInfo *parser.ServiceInfo) error {
	packageName := getPackageName(serviceInfo.GoPackage)
	fileName := fmt.Sprintf("%s_repo_impl.go", strings.ToLower(serviceInfo.Name))
	filePath := filepath.Join(outputDir, fileName)

	tplContent := `package {{.PackageName}}

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	{{range .ImportPaths}}
	"{{.}}"
	{{end}}
)

// {{.ServiceName}}RepoImpl is the implementation of {{.ServiceName}}Repo
type {{.ServiceName}}RepoImpl struct {
	data *Data
	log  *log.Helper
}

// New{{.ServiceName}}Repo creates a new {{.ServiceName}}Repo
func New{{.ServiceName}}Repo(data *Data, logger log.Logger) {{.ServiceName}}Repo {
	return &{{.ServiceName}}RepoImpl{
		data: data,
		log:  log.NewHelper(logger),
	}
}

{{range .Methods}}
// {{.Name}} implements {{$.ServiceName}}Repo
func (r *{{$.ServiceName}}RepoImpl) {{.Name}}(ctx context.Context, req *{{.InputName}}) (*{{.OutputName}}, error) {
	// Implement data access logic here
	r.log.WithContext(ctx).Infof("get {{.Name}} with %+v", req)
	return &{{.OutputName}}{}, nil
}
{{end}}
`

	tpl, err := template.New("dataRepoImpl").Parse(tplContent)
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

// generateDataModels generates the data models
func generateDataModels(outputDir string, serviceInfo *parser.ServiceInfo) error {
	packageName := getPackageName(serviceInfo.GoPackage)
	fileName := "model.go"
	filePath := filepath.Join(outputDir, fileName)

	tplContent := `package {{.PackageName}}

import (
	"time"
	"github.com/go-kratos/kratos/v2/errors"
)

// Data models for database operations

{{range .Messages}}
// {{.Name}} is the data model for {{.Name}}
type {{.Name}} struct {
	{{range .Fields}}
	{{.Name}} {{.Type}} {{if .TypeName}}// {{.TypeName}}{{end}}
	{{end}}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for {{.Name}}
func ({{.Name}}) TableName() string {
	return "{{toLower .Name}}"
}
{{end}}
`

	// Create custom template functions
	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
	}

	tpl, err := template.New("dataModels").Funcs(funcMap).Parse(tplContent)
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
