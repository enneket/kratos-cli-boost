package biz

var bizTemplate = `
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

{{ range .Entities }}
// {{ .Name }} 领域实体（业务核心数据结构）
type {{ .Name }} struct {
	{{- range .Fields }}
	{{ .FieldName }} {{ .FieldType }} // {{ .Comment }}
	{{- end }}
}
{{- end }}

type {{ .ServiceName }}Repo interface {
	{{- range .Methods }}
	// {{ .Comment }}
	{{ .MethodName }}(ctx context.Context{{- if .ParamName }}, {{ .ParamName }} {{ .ParamType }} {{ end }}) ({{ .ReturnType }}, error) 
	{{- end }}
}

type {{ .ServiceName }}UseCase struct {
	repo {{ .ServiceName }}Repo       // 依赖 Repo 接口（依赖抽象）
	log  *log.Helper                  // 日志组件
}

func New{{ .ServiceName }}UseCase(repo {{ .ServiceName }}Repo, logger log.Logger) *{{ .ServiceName }}UseCase {
	return &{{ .ServiceName }}UseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "usecase/{{ .ServiceName | toLower }}")),
	}
}

{{ range .Methods }}
// {{ .Comment }}
func (uc *{{ $.ServiceName }}UseCase) {{ .MethodName }}(ctx context.Context{{- if .ParamName }}, {{ .ParamName }} {{ .ParamType }} {{ end }}) ({{ .ReturnType }}, error) {
	data, err := uc.repo.{{ .MethodName }}(ctx{{- if .ParamName }}, {{ .ParamName }}{{ end }})
	if err != nil {
		uc.log.Errorf("{{ .MethodName }} repo operation failed: %v", err)
		return nil, err
	}

	return data, nil
}
{{- end }}
`
