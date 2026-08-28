package usecase

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/repository"
)

// NodeUsecase はノードに関するユースケース。
type NodeUsecase struct {
	repo repository.NodeRepository
}

// NewNodeUsecase は新しいNodeUsecaseを作成する。
func NewNodeUsecase(repo repository.NodeRepository) *NodeUsecase {
	return &NodeUsecase{repo: repo}
}

// List はノード一覧を取得する。
func (u *NodeUsecase) List(ctx context.Context, filter repository.NodeFilter) ([]model.Node, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	return u.repo.FindAll(ctx, filter)
}
