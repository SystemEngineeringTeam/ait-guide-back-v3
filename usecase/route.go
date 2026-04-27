package usecase

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/repository"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/service"
)

// RouteUsecase は経路探索に関するユースケース。
type RouteUsecase struct {
	nodeRepo  repository.NodeRepository
	routeRepo repository.RouteRepository
}

// NewRouteUsecase は新しいRouteUsecaseを作成する。
func NewRouteUsecase(nodeRepo repository.NodeRepository, routeRepo repository.RouteRepository) *RouteUsecase {
	return &RouteUsecase{nodeRepo: nodeRepo, routeRepo: routeRepo}
}

// SearchRoute は経路探索を実行する（新バージョン）。
func (u *RouteUsecase) SearchRoute(ctx context.Context, lat, lng float64, buildingID int, option model.RouteOption) (*model.Route, error) {
	// 最寄りノード検索
	sourceNode, err := u.nodeRepo.FindNearestByLatLng(ctx, lat, lng)
	if err != nil {
		return nil, err
	}
	if sourceNode == nil {
		return nil, service.ErrNodeNotFound
	}

	// 目的地の建物入口ノード検索
	targetNode, err := u.nodeRepo.FindEntranceByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	if targetNode == nil {
		return nil, service.ErrBuildingNotFound
	}

	// コスト重みを算出して経路探索
	weight := service.CostWeightFromOption(option)
	route, err := u.routeRepo.FindRoute(ctx, sourceNode.ID, targetNode.ID, option, weight)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return nil, service.ErrNoRouteFound
	}

	return route, nil
}

// SearchRouteLegacy は経路探索を実行する（旧バージョン）。
func (u *RouteUsecase) SearchRouteLegacy(ctx context.Context, lat, lng float64, endID int) (*model.Route, error) {
	// 最寄りノード検索
	sourceNode, err := u.nodeRepo.FindNearestByLatLng(ctx, lat, lng)
	if err != nil {
		return nil, err
	}
	if sourceNode == nil {
		return nil, service.ErrNodeNotFound
	}

	// デフォルトオプションで経路探索
	option := model.RouteOption{
		Level:  3,
		Stairs: true,
	}
	weight := service.CostWeightFromOption(option)
	route, err := u.routeRepo.FindRoute(ctx, sourceNode.ID, endID, option, weight)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return nil, service.ErrNoRouteFound
	}

	return route, nil
}
