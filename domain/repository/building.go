package repository

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
)

// BuildingRepository は建物データへのアクセスを提供する。
type BuildingRepository interface {
	FindAll(ctx context.Context, limit int, affiliation *string) ([]model.Building, error)
	FindByID(ctx context.Context, id int) (*model.Building, error)
}
