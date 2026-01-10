package compat

import (
	"proxy-system-backend/internal/service/core/proxy"
	"proxy-system-backend/internal/service/core/websocket"
)

// ShadowsocksService 兼容ShadowsocksService - 现在直接使用proxy.ShadowsocksService
type ShadowsocksService = proxy.ShadowsocksService

// NewShadowsocksService 创建兼容的Shadowsocks服务
func NewShadowsocksService(wsService *WebSocketService, pluginService *PluginService) *ShadowsocksService {
	// 直接创建service层的实例（使用兼容层的类型）
	serviceWSService := websocket.NewWebSocketService(websocket.WebSocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		PingInterval:    30 * 1000000000, // 30秒
	})

	// 直接使用兼容层的pluginService实例
	return proxy.NewShadowsocksService(serviceWSService, pluginService)
}

// 兼容性层 - ShadowsocksService现在直接使用service.ShadowsocksService
// 不需要额外的代理方法，直接使用service层的方法
