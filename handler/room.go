package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/repository"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/service"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/usecase"
	"github.com/gin-gonic/gin"
)

// RoomHandler は部屋関連のHTTPハンドラー。
type RoomHandler struct {
	uc *usecase.RoomUsecase
}

// NewRoomHandler は新しいRoomHandlerを作成する。
func NewRoomHandler(uc *usecase.RoomUsecase) *RoomHandler {
	return &RoomHandler{uc: uc}
}

// List は部屋一覧を取得する。
//
//	@Summary		部屋一覧取得
//	@Description	部屋の一覧を返す。建物ID・階数・部屋タイプでフィルタリング可能。
//	@Tags			rooms
//	@Produce		json
//	@Param			building_id	query		int				false	"建物IDでフィルタリング"
//	@Param			floor		query		int				false	"階数でフィルタリング"
//	@Param			room_type	query		string			false	"部屋タイプでフィルタリング（classroom, office, lab, lecture_hall）"
//	@Success		200			{object}	map[string]any	"部屋一覧"
//	@Failure		500			{object}	map[string]any	"サーバーエラー"
//	@Router			/rooms [get]
func (h *RoomHandler) List(c *gin.Context) {
	var filter repository.RoomFilter

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
	if v := c.Query("room_type"); v != "" {
		filter.RoomType = &v
	}

	rooms, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	type roomResp struct {
		RoomID     string  `json:"room_id"`
		Name       *string `json:"name,omitempty"`
		BuildingID int     `json:"building_id"`
		Floor      *int    `json:"floor,omitempty"`
		Capacity   *int    `json:"capacity,omitempty"`
		RoomType   *string `json:"room_type,omitempty"`
	}
	result := make([]roomResp, len(rooms))
	for i, r := range rooms {
		result[i] = roomResp{
			RoomID:     r.RoomID,
			Name:       r.Name,
			BuildingID: r.BuildingID,
			Floor:      r.Floor,
			Capacity:   r.Capacity,
			RoomType:   r.RoomType,
		}
	}

	respondSuccess(c, result)
}

// GetByRoomID は部屋情報を取得する。
//
//	@Summary		部屋情報取得
//	@Description	指定した部屋IDの詳細情報を返す
//	@Tags			rooms
//	@Produce		json
//	@Param			room_id	path		string			true	"部屋ID"
//	@Success		200		{object}	map[string]any	"部屋詳細"
//	@Failure		404		{object}	map[string]any	"部屋が見つからない"
//	@Failure		500		{object}	map[string]any	"サーバーエラー"
//	@Router			/rooms/{room_id} [get]
func (h *RoomHandler) GetByRoomID(c *gin.Context) {
	roomID := c.Param("room_id")

	rm, err := h.uc.GetByRoomID(c.Request.Context(), roomID)
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			respondError(c, http.StatusNotFound, "ROOM_NOT_FOUND", "部屋が見つかりません")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondSuccess(c, gin.H{
		"room_id":      rm.RoomID,
		"name":         rm.Name,
		"description":  rm.Description,
		"building_id":  rm.BuildingID,
		"floor":        rm.Floor,
		"capacity":     rm.Capacity,
		"room_type":    rm.RoomType,
		"door_node_id": rm.NodeID,
	})
}
