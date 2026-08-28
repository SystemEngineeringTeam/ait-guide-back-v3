// Package usecase はアプリケーションのユースケースを提供する。
package usecase

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/repository"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/service"
)

// BuildingUsecase は建物に関するユースケース。
type BuildingUsecase struct {
	repo repository.BuildingRepository
}

// NewBuildingUsecase は新しいBuildingUsecaseを作成する。
func NewBuildingUsecase(repo repository.BuildingRepository) *BuildingUsecase {
	return &BuildingUsecase{repo: repo}
}

// List は建物一覧を取得する。
func (u *BuildingUsecase) List(ctx context.Context, limit int, affiliation *string) ([]model.Building, error) {
	if limit <= 0 {
		limit = 100
	}
	return u.repo.FindAll(ctx, limit, affiliation)
}

// GetByID は建物情報を取得する。
func (u *BuildingUsecase) GetByID(ctx context.Context, id int) (*model.Building, error) {
	b, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, service.ErrBuildingNotFound
	}
	return b, nil
}
