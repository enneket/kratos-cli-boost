package data

var dataTemplate = `{{- /* go-kratos data 层模板：实现 domain Repo 接口 */ -}}
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	{{- if .UseLogger }}
	"github.com/go-kratos/kratos/v2/log"
	{{- end }}
	{{- if .DbPackage }}
	"{{ .DbPackage }}"
	{{- end }}
	{{- if .CachePackage }}
	"{{ .CachePackage }}"
	{{- end }}
	{{- if .ProtoPackage }}
	pb "{{ .ProtoPackage }}"
	{{- end }}

	"{{ .DomainPackage }}" // 依赖领域层的 Repo 接口和实体
)

// {{ .Service }}RepoImpl 实现 domain 层定义的 {{ .Service }}Repo 接口
type {{ .Service }}RepoImpl struct {
	{{- if .DbPackage }}
	// 数据库客户端（如 gorm.DB）
	db  *gorm.DB
	{{- end }}
	{{- if .CachePackage }}
	// 缓存客户端（如 redis.Client）
	cache *redis.Client
	{{- end }}
	{{- if .UseLogger }}
	// 日志组件
	log *log.Helper
	{{- end }}
}

// New{{ .Service }}Repo 创建 Repo 实例（依赖注入入口）
func New{{ .Service }}Repo(
	{{- if .DbPackage }}
	db *gorm.DB,
	{{- end }}
	{{- if .CachePackage }}
	cache *redis.Client,
	{{- end }}
	{{- if .UseLogger }}
	logger log.Logger,
	{{- end }}
) {{ .DomainPackage }}.{{ .Service }}Repo {
	return &{{ .Service }}RepoImpl{
		{{- if .DbPackage }}
		db:  db,
		{{- end }}
		{{- if .CachePackage }}
		cache: cache,
		{{- end }}
		{{- if .UseLogger }}
		log:  log.NewHelper(logger),
		{{- end }}
	}
}

{{- /* 遍历方法，生成 Repo 接口实现 */ -}}
{{- range .Methods }}
// {{ .MethodName }} 实现 {{ .RepoName }} 接口的 {{ .MethodName }} 方法
// 职责：数据访问逻辑（数据库 CRUD / 缓存操作）
func (r *{{ $.Service }}RepoImpl) {{ .MethodName }}(ctx context.Context, req {{ .ParamType }}) ({{ .ReturnType }}, error) {
	{{- if $.UseLogger }}
	r.log.Infof("{{ .MethodName }} data layer start, req: %+v", req)
	{{- end }}

	// 1. （可选）缓存查询：优先从缓存获取数据
	// cacheKey := fmt.Sprintf("{{ $.Service }}:%s", req.Id)
	// var cacheData {{ .ReturnType }}
	// if err := r.cache.Get(ctx, cacheKey).Scan(&cacheData); err == nil {
	// 	r.log.Infof("{{ .MethodName }} hit cache, key: %s", cacheKey)
	// 	return cacheData, nil
	// }

	// 2. 数据库操作：根据业务逻辑实现 CRUD
	// 示例（GORM）：
	// var dbModel {{ $.Service }}Model // 数据层模型（与数据库表对应）
	// // domain 实体 → 数据层模型（补充转换逻辑）
	// dbModel = convertDomainEntityToDbModel(req)
	// if err := r.db.WithContext(ctx).Create(&dbModel).Error; err != nil {
	// 	r.log.Errorf("{{ .MethodName }} db create failed: %v", err)
	// 	return {{ .ReturnType }}{}, err
	// }

	// 3. （可选）写入缓存：更新缓存数据
	// resultEntity := convertDbModelToDomainEntity(dbModel)
	// if err := r.cache.Set(ctx, cacheKey, resultEntity, time.Hour).Err(); err != nil {
	// 	r.log.Warnf("{{ .MethodName }} set cache failed: %v", err)
	// }

	// 4. 数据层模型 → domain 实体（返回给 biz 层）
	result := {{ .ReturnType }}{
		// 补充字段映射逻辑，示例：
		// Id:   dbModel.Id,
		// Name: dbModel.Name,
		// ...
	}

	{{- if $.UseLogger }}
	r.log.Infof("{{ .MethodName }} data layer success, result: %+v", result)
	{{- end }}
	return result, nil
}
{{- end }}

{{- /* 预留实体转换辅助函数（按需补充） */ -}}
// convertDomainEntityToDbModel domain 实体 → 数据层模型
// func convertDomainEntityToDbModel(entity {{ .DomainPackage }}.{{ .Service }}Entity) {{ .Service }}Model {
// 	return {{ .Service }}Model{
// 		Id:   entity.Id,
// 		Name: entity.Name,
// 		// ... 其他字段映射
// 	}
// }

// convertDbModelToDomainEntity 数据层模型 → domain 实体
// func convertDbModelToDomainEntity(model {{ .Service }}Model) {{ .DomainPackage }}.{{ .Service }}Entity {
// 	return {{ .DomainPackage }}.{{ .Service }}Entity{
// 		Id:   model.Id,
// 		Name: model.Name,
// 		// ... 其他字段映射
// 	}
// }

{{- /* 预留数据层模型定义（与数据库表对应） */ -}}
// {{ .Service }}Model 数据层模型（示例，根据实际数据库表结构调整）
// type {{ .Service }}Model struct {
// 	Id        string    ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
// 	Name      string    ` + "`gorm:\"column:name\" json:\"name\"`" + `
// 	CreatedAt time.Time ` + "`gorm:\"autoCreateTime\" json:\"created_at\"`" + `
// 	UpdatedAt time.Time ` + "`gorm:\"autoUpdateTime\" json:\"updated_at\"`" + `
// }

// TableName 定义数据库表名
// func (m {{ .Service }}Model) TableName() string {
// 	return "{{ strings.ToLower .Service }}"
// }
`
