package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler はヘルスチェックハンドラー。
type HealthHandler struct{}

// NewHealthHandler は新しいHealthHandlerを作成する。
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health はヘルスチェックを行う。
//
//	@Summary		ヘルスチェック
//	@Description	APIの健全性を確認する
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	map[string]string	"status と timestamp を返す"
//	@Router			/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
