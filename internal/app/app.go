package app

import (
	"net"
	"proxy-system-backend/internal/modules/proxy"
	"proxy-system-backend/internal/modules/shadowsocks"
	"proxy-system-backend/internal/modules/shared"
	"proxy-system-backend/internal/traffic"

	"sync"
)

type App struct {
	mu        sync.RWMutex
	listeners []func(Event)
	proxyMgr  *ProxyManager
}

func New() *App {
	return &App{
		proxyMgr:  NewProxyManager(),
		listeners: make([]func(Event), 0),
	}
}

// Subscribe 订阅事件（websocket / logger / metrics 都可以）
func (a *App) Subscribe(fn func(Event)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.listeners = append(a.listeners, fn)
}

// Emit domain 调用这个
func (a *App) Emit(e Event) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, fn := range a.listeners {
		fn(e)
	}
}

// tmp
func (a *App) StartProxy(cfg proxy.Config) error {
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}
	proxyID := shared.GenerateConnID()
	c, err := cfg.BuildCipher()
	server := shadowsocks.NewServer(
		ln,
		c,
		DefaultDirectDialer(),
		func(connID string) traffic.TrafficHook {
			return a.newTrafficHook(proxyID, connID)
		},
	)

	return a.proxyMgr.StartProxy(proxyID, server)

}
func (a *App) newTrafficHook(proxyID, connID string) traffic.TrafficHook {
	return &proxyTrafficHook{
		app:     a,
		proxyID: proxyID,
		connID:  connID,
	}
}
