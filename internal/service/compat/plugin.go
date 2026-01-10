package compat

import (
	"proxy-system-backend/internal/service/plugin"
	"proxy-system-backend/internal/types"
)

// PluginService 兼容PluginService - 直接使用plugin.PluginService
type PluginService = plugin.PluginService

// NewPluginService 创建兼容的插件服务 - 直接使用plugin.NewPluginService
func NewPluginService(config types.PluginManagerConfig) *PluginService {
	return plugin.NewPluginService(config)
}
