package data

var dataTemplate = `{{- /* go-kratos data 层模板：实现 domain Repo 接口 */ -}}
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"{{ .UseCasePackage }}" // 依赖领域层的 Repo 接口和实体
)

// {{ .Service }}Repo 实现 biz 层定义的 {{ .Service }}Repo 接口
type {{ .Service }}Repo struct {

}

// New{{ .Service }}Repo 创建 Repo 实例（依赖注入入口）
func New{{ .Service }}Repo() biz.{{ .Service }}Repo {
	return &{{ .Service }}Repo{}
}

{{- /* 遍历方法，生成 Repo 接口实现 */ -}}
{{- range .Methods }}
func (r *{{ $.Service }}Repo) {{ .MethodName }}(ctx context.Context, req {{ .ParamType }}) ({{ .ReturnType }}, error) {
	panic("unimplemented")
}
{{- end }}
`
