package repository

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
)

// NodeRepository はノードデータへのアクセスを提供する。
type NodeRepository interface {
	FindAll(ctx context.Context, filter NodeFilter) ([]model.Node, error)
	FindNearestByLatLng(ctx context.Context, lat, lng float64) (*model.Node, error)
	// FindNearestTargetByBuildingID は指定した建物IDを持つentrance/facilityノードのうち、
	// 指定した緯度経度から直線距離が最も近いものを返す。entranceを持たない施設ではfacilityが対象になる。
	FindNearestTargetByBuildingID(ctx context.Context, buildingID int, lat, lng float64) (*model.Node, error)
}

// NodeFilter はノード検索のフィルタ条件。
type NodeFilter struct {
	NodeID     *string
	NodeType   *model.NodeType
	BuildingID *int
	Floor      *int
	Limit      int
}
