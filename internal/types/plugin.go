package types

import (
	"time"
)

// PluginInfo 插件信息结构
type PluginInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Status      string    `json:"status"` // "active", "inactive", "error"
	Description string    `json:"description"`
}

// PluginUploadRequest 插件上传请求
type PluginUploadRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PluginResponse 插件操作响应
type PluginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PluginListResponse 插件列表响应
type PluginListResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    []PluginInfo `json:"data"`
	Total   int64        `json:"total"`
}

// PluginManagerConfig 插件管理器配置
type PluginManagerConfig struct {
	PluginDir string `json:"plugin_dir"`
	MaxSize   int64  `json:"max_size"` // 插件文件最大大小（字节）
}
