package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"proxy-system-backend/internal/handler"
	"proxy-system-backend/internal/middleware"
	"proxy-system-backend/internal/service/compat"
	"proxy-system-backend/internal/service/core/websocket"
	"proxy-system-backend/internal/types"
)

func main() {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	// 初始化插件服务（使用兼容性包装器）
	pluginConfig := types.PluginManagerConfig{
		PluginDir: "./plugins",
		MaxSize:   50 * 1024 * 1024, // 50MB
	}
	pluginService := compat.NewPluginService(pluginConfig)
	if err := pluginService.Initialize(); err != nil {
		log.Printf("Warning: failed to initialize plugin service: %v", err)
	}

	// 初始化WebSocket服务
	wsService := compat.NewWebSocketService()

	// 启动 WebSocket 服务
	go wsService.Run()

	// 创建Shadowsocks服务（包含集成过滤功能）
	shadowsocksService := compat.NewShadowsocksService(wsService, pluginService)

	// 启动 WebSocket 服务
	go wsService.Run()

	api := r.Group("/api")
	{
		handler.SetupWebSocketRoutes(api, wsService)
		 
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 使用 defer 确保程序退出时清理资源
	defer func() {
		if pluginService != nil {
			// 插件服务会在垃圾回收时自动关闭数据库连接
			log.Println("Cleaning up resources...")
		}
	}()

	log.Println("Starting HTTP server on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}
