package handler

import (
	"proxy-system-backend/internal/app"
)

type FilterHandler struct {
	app *app.App
}

//func (h *FilterHandler) ListRules(c *gin.Context) {
//	//rules, err := h.app.FilterEngine().
//	if err != nil {
//		c.JSON(500, gin.H{"success": false, "error": err.Error()})
//		return
//	}
//
//	c.JSON(200, gin.H{
//		"success": true,
//		"data":    rules,
//	})
//}
