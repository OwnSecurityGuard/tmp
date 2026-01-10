package app

import (
	websocket2 "proxy-system-backend/internal/modules/websocket"
	"time"
)

type Wiring struct {
	App   *App
	WSHub *websocket2.Hub
}

func Wire() *Wiring {
	// 1️⃣ websocket hub
	wsHub := websocket2.NewHub(websocket2.Config{
		PingInterval: 10 * time.Second,
	})
	go wsHub.Run()

	// 2️⃣ app core
	app := New()

	// 3️⃣ websocket notifier
	wsNotifier := NewWSNotifier(wsHub)
	app.Subscribe(wsNotifier.HandleEvent)

	//db, _ := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	//db.AutoMigrate(&filterstore.RuleModel{})
	//
	//repo := filterstore.NewSQLiteRepo(db)
	//engine := filter.NewEngine()
	//loader := NewFilterLoader(repo, engine)
	//_ = loader.Load(context.Background())

	return &Wiring{
		App:   app,
		WSHub: wsHub,
		//Shadowsocks: ss,
	}
}
