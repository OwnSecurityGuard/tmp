package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"proxy-system-backend/internal/app"
)

type PluginHandler struct {
	app *app.App
}

func NewPluginHandler(a *app.App) *PluginHandler {
	return &PluginHandler{app: a}
}

type RegisterPluginReq struct {
	Name string `json:"name" binding:"required"`
	Path string `json:"path" binding:"required"`
}

func (h *PluginHandler) Register(c *gin.Context) {
	var req RegisterPluginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.app.PluginMgr().Register(req.Name, req.Path); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func (h *PluginHandler) Load(c *gin.Context) {
	name := c.Param("name")

	if err := h.app.PluginMgr().Load(name); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func (h *PluginHandler) Unload(c *gin.Context) {
	name := c.Param("name")

	if err := h.app.PluginMgr().Unload(name); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}
func (h *PluginHandler) List(c *gin.Context) {
	data, _ := h.app.PluginMgr().List()

	c.JSON(200, gin.H{
		"plugins": data,
	})
}
func (h *PluginHandler) Get(c *gin.Context) {
	name := c.Param("name")

	info, err := h.app.PluginMgr().Get(name)
	if err != nil {
		c.JSON(404, gin.H{"error": "plugin not found"})
		return
	}

	c.JSON(200, info)
}

const (
	PluginDir = "E:\\reply\\backend\\data\\plugins"
)

func (h *PluginHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required",
		})
		return
	}

	// 插件名：优先用 form 里的 name，否则用文件名
	name := c.PostForm("name")
	if name == "" {
		name = file.Filename
	}

	// 最终目录：/data/plugins/<name>/
	baseDir := PluginDir
	pluginDir := filepath.Join(baseDir, name)

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 最终路径
	dstPath := filepath.Join(pluginDir, file.Filename)

	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 确保可执行（Linux / macOS）
	_ = os.Chmod(dstPath, 0755)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"path":    dstPath,
	})
}
