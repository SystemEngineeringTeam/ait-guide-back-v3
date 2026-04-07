package repository

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
)

// RouteRepository は経路探索を提供する。
type RouteRepository interface {
	FindRoute(ctx context.Context, sourceNodeID, targetNodeID int, option model.RouteOption, weight model.CostWeight) (*model.Route, error)
}
