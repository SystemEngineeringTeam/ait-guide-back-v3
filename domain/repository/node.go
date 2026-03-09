package repository

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
)

// NodeRepository はノードデータへのアクセスを提供する。
type NodeRepository interface {
	FindAll(ctx context.Context, filter NodeFilter) ([]model.Node, error)
	FindNearestByLatLng(ctx context.Context, lat, lng float64) (*model.Node, error)
	FindEntranceByBuildingID(ctx context.Context, buildingID int) (*model.Node, error)
}

// NodeFilter はノード検索のフィルタ条件。
type NodeFilter struct {
	NodeID     *string
	NodeType   *model.NodeType
	BuildingID *int
	Floor      *int
	Limit      int
}
