package biz

var bizTemplate = `
{{- /* go-kratos biz 层模板，包含业务接口与实现 */ -}}
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	
	pb "{{ .ProtoPackage }}" // proto 生成的 PB 包路径
)

// {{ .Service }}UseCase 定义业务接口（包含核心业务方法）
type {{ .Service }}UseCase interface {
	{{- range .Methods }}
	// {{ .Name }} 处理 {{ .Name }} 业务逻辑
	{{ .Name }}(ctx context.Context, req *pb.{{ .Request }}) (*pb.{{ .Reply }}, error)
	{{- end }}
}

// {{ .Service }}UseCaseImpl 业务接口实现
type {{ .Service }}UseCaseImpl struct {
	// 依赖注入：领域仓库（操作领域实体）
	repo {{ .DomainPackage }}.{{ .Service }}Repo
	// 日志组件
	log *log.Helper
}

// New{{ .Service }}UseCase 创建业务实例
func New{{ .Service }}UseCase(repo {{ .DomainPackage }}.{{ .Service }}Repo, logger log.Logger) {{ .Service }}UseCase {
	return &{{ .Service }}UseCaseImpl{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

{{- /* 遍历方法生成业务实现框架 */ -}}
{{- range .Methods }}
// {{ .Name }} 实现 {{ .Name }} 业务逻辑
func (uc *{{ .Service }}UseCaseImpl) {{ .Name }}(ctx context.Context, req *pb.{{ .Request }}) (*pb.{{ .Reply }}, error) {
	// 1. 参数校验
	if err := uc.validate{{ .Name }}Req(req); err != nil {
		uc.log.Errorf("{{ .Name }} request validate failed: %v", err)
		return nil, err
	}

	// 2. 转换 PB 请求为领域实体（如需）
	// entity := {{ .DomainPackage }}.Convert{{ .Request }}ToEntity(req)

	// 3. 调用领域仓库处理核心业务（示例）
	// resultEntity, err := uc.repo.{{ .Name }}(ctx, entity)
	// if err != nil {
	// 	uc.log.Errorf("{{ .Name }} repo failed: %v", err)
	// 	return nil, errors.FromError(err)
	// }

	// 4. 转换领域实体为 PB 响应（如需）
	resp := &pb.{{ .Reply }}{
		// 填充响应字段，示例：
		// Id: resultEntity.Id,
		// Message: "success",
	}

	uc.log.Infof("{{ .Name }} business handled successfully, req: %+v, resp: %+v", req, resp)
	return resp, nil
}

// validate{{ .Name }}Req 校验 {{ .Name }} 请求参数
func (uc *{{ .Service }}UseCaseImpl) validate{{ .Name }}Req(req *pb.{{ .Request }}) error {
	if req == nil {
		return errors.BadRequest("INVALID_REQUEST", "request is nil")
	}
	// 补充具体参数校验逻辑，例如：
	// if req.Id == "" {
	// 	return errors.BadRequest("INVALID_PARAM", "id is required")
	// }
	return nil
}
{{- end }}
`
