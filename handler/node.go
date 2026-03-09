package handler

import (
	"net/http"
	"strconv"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/repository"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/usecase"
	"github.com/gin-gonic/gin"
)

// NodeHandler はノード関連のHTTPハンドラー。
type NodeHandler struct {
	uc *usecase.NodeUsecase
}

// NewNodeHandler は新しいNodeHandlerを作成する。
func NewNodeHandler(uc *usecase.NodeUsecase) *NodeHandler {
	return &NodeHandler{uc: uc}
}

func (h *NodeHandler) List(c *gin.Context) {
	var filter repository.NodeFilter
	filter.Limit = 100

	if v := c.Query("node_id"); v != "" {
		filter.NodeID = &v
	}
	if v := c.Query("node_type"); v != "" {
		nt := model.NodeType(v)
		filter.NodeType = &nt
	}
	if v := c.Query("building_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			filter.BuildingID = &id
		}
	}
	if v := c.Query("floor"); v != "" {
		if f, err := strconv.Atoi(v); err == nil {
			filter.Floor = &f
		}
	}
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	nodes, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	type nodeResp struct {
		NodeID     string       `json:"node_id"`
		Lat        float64      `json:"lat"`
		Lng        float64      `json:"lng"`
		NodeType   model.NodeType `json:"node_type"`
		BuildingID *int         `json:"building_id,omitempty"`
		Floor      *int         `json:"floor,omitempty"`
	}
	result := make([]nodeResp, len(nodes))
	for i, n := range nodes {
		result[i] = nodeResp{
			NodeID:     n.NodeID,
			Lat:        n.Lat,
			Lng:        n.Lng,
			NodeType:   n.NodeType,
			BuildingID: n.BuildingID,
			Floor:      n.Floor,
		}
	}

	respondSuccess(c, result)
}
