package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"proxy-system-backend/internal/app"
	"proxy-system-backend/internal/modules/proxy"
	"time"
)

type ProxyHandler struct {
	app *app.App
}

func NewProxyHandler(app *app.App) *ProxyHandler {
	return &ProxyHandler{app: app}
}

func (h *ProxyHandler) StartProxy(c *gin.Context) {
	var req proxy.Config
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 补生命周期字段
	now := time.Now().Unix()
	req.Enabled = true
	req.CreatedAt = now
	req.UpdatedAt = now

	if err := h.app.StartProxy(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"proxy_id": req.ID,
	})
}
