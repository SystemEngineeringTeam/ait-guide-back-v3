package handler

import (
	"github.com/gin-gonic/gin"
)

// Router はAPIルーターを構成する。
func Router(
	healthH *HealthHandler,
	buildingH *BuildingHandler,
	roomH *RoomHandler,
	nodeH *NodeHandler,
	routeH *RouteHandler,
) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/health", healthH.Health)

		// 経路探索
		api.GET("/get/route", routeH.Legacy)
		api.GET("/routes/search/:building_id", routeH.Search)

		// 建物
		api.GET("/buildings", buildingH.List)
		api.GET("/buildings/:building_id", buildingH.GetByID)

		// 部屋
		api.GET("/rooms", roomH.List)
		api.GET("/rooms/:room_id", roomH.GetByRoomID)

		// ノード
		api.GET("/nodes", nodeH.List)
	}

	return r
}
