package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/service"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/usecase"
	"github.com/gin-gonic/gin"
)

const (
	defaultLat = 35.181531
	defaultLng = 137.109509
)

// RouteHandler は経路探索関連のHTTPハンドラー。
type RouteHandler struct {
	uc *usecase.RouteUsecase
}

// NewRouteHandler は新しいRouteHandlerを作成する。
func NewRouteHandler(uc *usecase.RouteUsecase) *RouteHandler {
	return &RouteHandler{uc: uc}
}

// Search は経路探索（新バージョン）のハンドラー。
func (h *RouteHandler) Search(c *gin.Context) {
	buildingID, err := strconv.Atoi(c.Param("building_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_PARAMETERS", "building_idが不正です")
		return
	}

	lat := parseFloat(c.Query("lat"), defaultLat)
	lng := parseFloat(c.Query("lng"), defaultLng)

	option := model.RouteOption{
		Level:      parseInt(c.Query("level"), 3),
		Stairs:     parseBoolParam(c.Query("stairs"), true),
		Accessible: parseBoolParam(c.Query("accessible"), false),
		Indoor:     parseBoolParam(c.Query("indoor"), false),
	}

	route, err := h.uc.SearchRoute(c.Request.Context(), lat, lng, buildingID, option)
	if err != nil {
		handleRouteError(c, err)
		return
	}

	type pointResp struct {
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
		IsIndoor bool    `json:"is_indoor"`
	}
	points := make([]pointResp, len(route.Points))
	for i, p := range route.Points {
		points[i] = pointResp{Lat: p.Lat, Lng: p.Lng, IsIndoor: p.IsIndoor}
	}

	respondSuccess(c, gin.H{
		"route":          points,
		"total_distance": route.TotalDistance,
		"total_cost":     route.TotalCost,
	})
}

// Legacy は経路探索（旧バージョン）のハンドラー。
func (h *RouteHandler) Legacy(c *gin.Context) {
	lat := parseFloat(c.Query("lat"), defaultLat)
	lng := parseFloat(c.Query("lng"), defaultLng)
	endNodeID := c.DefaultQuery("end", "1")

	route, err := h.uc.SearchRouteLegacy(c.Request.Context(), lat, lng, endNodeID)
	if err != nil {
		handleRouteError(c, err)
		return
	}

	type pointResp struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	points := make([]pointResp, len(route.Points))
	for i, p := range route.Points {
		points[i] = pointResp{Lat: p.Lat, Lng: p.Lng}
	}

	c.JSON(http.StatusOK, gin.H{
		"route": points,
	})
}

func handleRouteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNodeNotFound):
		respondError(c, http.StatusNotFound, "NODE_NOT_FOUND", "ノードが見つかりません")
	case errors.Is(err, service.ErrBuildingNotFound):
		respondError(c, http.StatusNotFound, "BUILDING_NOT_FOUND", "建物が見つかりません")
	case errors.Is(err, service.ErrNoRouteFound):
		respondError(c, http.StatusNotFound, "NO_ROUTE_FOUND", "経路が見つかりません")
	default:
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseBoolParam(s string, defaultVal bool) bool {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return v
}
