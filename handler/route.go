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
//
//	@Summary		経路探索
//	@Description	複数の重みパラメータとフィルタリングオプションを使用した経路探索
//	@Tags			routes
//	@Produce		json
//	@Param			building_id	path		int				true	"目的地の建物ID"
//	@Param			lat			query		number			false	"現在地の緯度"		default(35.181531)
//	@Param			lng			query		number			false	"現在地の経度"		default(137.109509)
//	@Param			level		query		int				false	"道の優先度（0:距離のみ, 1:大通り優先〜5:脇道優先）"	default(3)
//	@Param			stairs		query		boolean			false	"階段を含む経路を許可するか"	default(true)
//	@Param			accessible	query		boolean			false	"バリアフリー経路を使用するか"	default(false)
//	@Param			indoor		query		boolean			false	"屋内優先か"		default(false)
//	@Success		200			{object}	map[string]any	"経路情報"
//	@Failure		400			{object}	map[string]any	"パラメータ不正"
//	@Failure		404			{object}	map[string]any	"ノード/建物/経路が見つからない"
//	@Failure		500			{object}	map[string]any	"サーバーエラー"
//	@Router			/routes/search/{building_id} [get]
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
//
//	@Summary		経路探索（レガシー）
//	@Description	旧バージョンとの互換性のためのエンドポイント。距離のみを重みとして使用。
//	@Tags			routes
//	@Produce		json
//	@Param			lat	query		number			false	"現在地の緯度"		default(35.181531)
//	@Param			lng	query		number			false	"現在地の経度"		default(137.109509)
//	@Param			end	query		int				false	"目的地のノードID（nodes.id）"	default(1)
//	@Success		200	{object}	map[string]any	"経路情報"
//	@Failure		404	{object}	map[string]any	"ノード/経路が見つからない"
//	@Failure		500	{object}	map[string]any	"サーバーエラー"
//	@Router			/get/route [get]
func (h *RouteHandler) Legacy(c *gin.Context) {
	lat := parseFloat(c.Query("lat"), defaultLat)
	lng := parseFloat(c.Query("lng"), defaultLng)
	endID := parseInt(c.Query("end"), 1)

	route, err := h.uc.SearchRouteLegacy(c.Request.Context(), lat, lng, endID)
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
