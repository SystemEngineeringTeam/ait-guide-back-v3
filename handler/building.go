package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/service"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/usecase"
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
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	result := make([]buildingResp, len(buildings))
	for i, b := range buildings {
		result[i] = buildingResp{ID: b.ID, Name: b.Name}
	}

	respondSuccess(c, result)
}

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
		"name":        b.Name,
		"description": b.Description,
		"photos":      photos,
	})
}
