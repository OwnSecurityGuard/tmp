package app

import (
	"fmt"
	"proxy-system-backend/internal/modules/shadowsocks"
	"sync"
)

type ProxyManager struct {
	mu      sync.Mutex
	servers map[string]*shadowsocks.Server
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		servers: make(map[string]*shadowsocks.Server),
	}
}

func (pm *ProxyManager) StartProxy(
	id string,
	server *shadowsocks.Server,
) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, ok := pm.servers[id]; ok {
		return fmt.Errorf("proxy %s already running", id)
	}

	pm.servers[id] = server

	go func() {
		if err := server.Serve(); err != nil {
			// TODO: log / notify ws
		}
	}()

	return nil
}

func (pm *ProxyManager) StopProxy(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	srv, ok := pm.servers[id]
	if !ok {
		return fmt.Errorf("proxy %s not found", id)
	}

	_ = srv.Close() // 你可以实现 Close
	delete(pm.servers, id)
	return nil
}
