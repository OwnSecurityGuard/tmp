package main

import (
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"proxy-system-backend/internal/app"
	"proxy-system-backend/internal/handler"
	"proxy-system-backend/internal/modules/plugin"
	"proxy-system-backend/internal/modules/websocket"
	pluginstore "proxy-system-backend/internal/storage/plugin"

	"time"
)

func main() {
	appCore := app.New()

	// ===== 2️⃣ WebSocket Hub =====
	hub := websocket.NewHub(websocket.Config{
		PingInterval: 30 * time.Second,
	})
	go hub.Run()

	// ===== 3️⃣ App → WS =====
	appCore.Subscribe(func(e app.Event) {
		// 统一事件 → JSON

		data, err := json.Marshal(e)
		if err != nil {
			return
		}
		hub.Broadcast(data)
	})

	// ===== 4️⃣ HTTP =====
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // 前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// WebSocket
	r.GET("/api/ws", hub.Handle)
	db, _ := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})

	// API
	proxyHandler := handler.NewProxyHandler(appCore)
	pluginHandler := handler.NewPluginHandler(appCore)
	pluginMgr := plugin.NewManager()

	db.AutoMigrate(pluginstore.PluginModel{})
	pluginRepo := pluginstore.NewPluginRepo(db)

	pluginSvc := app.NewPluginService(pluginRepo, pluginMgr)
	//appCore.
	appCore.SetPluginMgr(pluginSvc)
	if err := pluginSvc.Bootstrap(); err != nil {
		log.Println("plugin bootstrap failed:", err)
		return
	}

	api := r.Group("/api")
	{
		api.POST("/proxy/start", proxyHandler.StartProxy)
	}
	plugins := api.Group("/plugins")
	{
		plugins.POST("", pluginHandler.Register)
		plugins.GET("", pluginHandler.List)
		plugins.GET("/:name", pluginHandler.Get)
		plugins.POST("/:name/load", pluginHandler.Load)
		plugins.POST("/:name/unload", pluginHandler.Unload)
		plugins.POST("/upload", pluginHandler.Upload)
	}
	// ===== 5️⃣ Start =====
	addr := ":8081"
	log.Println("🚀 server listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
