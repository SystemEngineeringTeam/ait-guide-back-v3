package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/service"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/usecase"
	"github.com/gin-gonic/gin"
)

// BuildingHandler は建物関連のHTTPハンドラー。
type BuildingHandler struct {
	uc *usecase.BuildingUsecase
}

// NewBuildingHandler は新しいBuildingHandlerを作成する。
func NewBuildingHandler(uc *usecase.BuildingUsecase) *BuildingHandler {
	return &BuildingHandler{uc: uc}
}

// List は建物一覧を取得する。
//
//	@Summary		建物一覧取得
//	@Description	建物の一覧を返す。所属でフィルタリング可能。
//	@Tags			buildings
//	@Produce		json
//	@Param			limit		query		int		false	"返却する建物の最大数"	default(100)
//	@Param			affiliation	query		string	false	"所属でフィルタリング"
//	@Success		200			{object}	map[string]any	"建物一覧"
//	@Failure		500			{object}	map[string]any	"サーバーエラー"
//	@Router			/buildings [get]
func (h *BuildingHandler) List(c *gin.Context) {
	limit := 100
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			limit = l
		}
	}

	var affiliation *string
	if v := c.Query("affiliation"); v != "" {
		affiliation = &v
	}

	buildings, err := h.uc.List(c.Request.Context(), limit, affiliation)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	type buildingResp struct {
		ID   int     `json:"id"`
		Key  *string `json:"key,omitempty"`
		Name string  `json:"name"`
	}
	result := make([]buildingResp, len(buildings))
	for i, b := range buildings {
		result[i] = buildingResp{ID: b.ID, Key: b.Key, Name: b.Name}
	}

	respondSuccess(c, result)
}

// GetByID は建物情報を取得する。
//
//	@Summary		建物情報取得
//	@Description	指定した建物IDの詳細情報（写真含む）を返す
//	@Tags			buildings
//	@Produce		json
//	@Param			building_id	path		int				true	"建物ID"
//	@Success		200			{object}	map[string]any	"建物詳細"
//	@Failure		400			{object}	map[string]any	"パラメータ不正"
//	@Failure		404			{object}	map[string]any	"建物が見つからない"
//	@Failure		500			{object}	map[string]any	"サーバーエラー"
//	@Router			/buildings/{building_id} [get]
func (h *BuildingHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("building_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_PARAMETERS", "building_idが不正です")
		return
	}

	b, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrBuildingNotFound) {
			respondError(c, http.StatusNotFound, "BUILDING_NOT_FOUND", "建物が見つかりません")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	type photoResp struct {
		URL     string  `json:"url"`
		Caption *string `json:"caption,omitempty"`
	}
	photos := make([]photoResp, len(b.Photos))
	for i, p := range b.Photos {
		photos[i] = photoResp{URL: p.URL, Caption: p.Caption}
	}

	respondSuccess(c, gin.H{
		"id":          b.ID,
		"key":         b.Key,
		"name":        b.Name,
		"description": b.Description,
		"photos":      photos,
	})
}
