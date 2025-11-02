
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)


// CreateExample 领域实体（业务核心数据结构）
type CreateExample struct {
}
// UpdateExample 领域实体（业务核心数据结构）
type UpdateExample struct {
}
// DeleteExample 领域实体（业务核心数据结构）
type DeleteExample struct {
}
// GetExample 领域实体（业务核心数据结构）
type GetExample struct {
}
// ListExample 领域实体（业务核心数据结构）
type ListExample struct {
}

type ExampleRepo interface {
	// 
	CreateExample(ctx context.Context, createExample *CreateExample ) (*CreateExample, error)
	// 
	UpdateExample(ctx context.Context, updateExample *UpdateExample ) (*UpdateExample, error)
	// 
	DeleteExample(ctx context.Context, deleteExample *DeleteExample ) (*DeleteExample, error)
	// 
	GetExample(ctx context.Context, getExample *GetExample ) (*GetExample, error)
	// 
	ListExample(ctx context.Context, listExample *ListExample ) (*ListExample, error)
}

type ExampleUseCase struct {
	repo ExampleRepo       // 依赖 Repo 接口（依赖抽象）
	log  *log.Helper                  // 日志组件
}

func NewExampleUseCase(repo ExampleRepo, logger log.Logger) *ExampleUseCase {
	return &ExampleUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "usecase/example")),
	}
}


// 
func (uc *ExampleUseCase) CreateExample(ctx context.Context, createExample *CreateExample ) (*CreateExample, error) {
	data, err := uc.repo.CreateExample(ctx, createExample)
	if err != nil {
		uc.log.Errorf("CreateExample repo operation failed: %v", err)
		return nil, err
	}

	return data, nil
}
// 
func (uc *ExampleUseCase) UpdateExample(ctx context.Context, updateExample *UpdateExample ) (*UpdateExample, error) {
	data, err := uc.repo.UpdateExample(ctx, updateExample)
	if err != nil {
		uc.log.Errorf("UpdateExample repo operation failed: %v", err)
		return nil, err
	}

	return data, nil
}
// 
func (uc *ExampleUseCase) DeleteExample(ctx context.Context, deleteExample *DeleteExample ) (*DeleteExample, error) {
	data, err := uc.repo.DeleteExample(ctx, deleteExample)
	if err != nil {
		uc.log.Errorf("DeleteExample repo operation failed: %v", err)
		return nil, err
	}

	return data, nil
}
// 
func (uc *ExampleUseCase) GetExample(ctx context.Context, getExample *GetExample ) (*GetExample, error) {
	data, err := uc.repo.GetExample(ctx, getExample)
	if err != nil {
		uc.log.Errorf("GetExample repo operation failed: %v", err)
		return nil, err
	}

	return data, nil
}
// 
func (uc *ExampleUseCase) ListExample(ctx context.Context, listExample *ListExample ) (*ListExample, error) {
	data, err := uc.repo.ListExample(ctx, listExample)
	if err != nil {
		uc.log.Errorf("ListExample repo operation failed: %v", err)
		return nil, err
	}

	return data, nil
}
