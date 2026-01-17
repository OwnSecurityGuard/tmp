package app

import (
	"fmt"
	"net"
	"proxy-system-backend/internal/modules/filter"
	"proxy-system-backend/internal/modules/proxy"
	"proxy-system-backend/internal/modules/shadowsocks"
	"proxy-system-backend/internal/modules/shared"
	"proxy-system-backend/internal/traffic"

	"sync"
)

type App struct {
	mu           sync.RWMutex
	listeners    []func(Event)
	proxyMgr     *ProxyManager
	filterEngine *filter.Engine
	pluginMgr    *PluginService
}

func New() *App {
	return &App{
		proxyMgr:     NewProxyManager(),
		listeners:    make([]func(Event), 0),
		filterEngine: filter.NewEngine(),
		//pluginMgr :NewPluginService(),
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

type StartProxyOptions struct {
	FilterEngine *filter.Engine
}

func (a *App) StartProxy(cfg proxy.Config) error {
	// 1️⃣ 监听端口
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}
	fmt.Println(cfg.ListenAddr)
	// 2️⃣ proxy 实例 ID（不是 connID）
	proxyID := shared.GenerateConnID()

	err = a.PluginMgr().Load("test")
	fmt.Println("laod ", err)

	// 3️⃣ 构建加密器
	c, err := cfg.BuildCipher()
	if err != nil {
		return err
	}
	var sf *SimpleFilter
	if len(cfg.BlockIPs) > 0 || len(cfg.BlockPorts) > 0 {
		sf, err = NewSimpleFilter(cfg.BlockIPs, cfg.BlockPorts)
		if err != nil {
			return err
		}
	}
	//if cfg.EnableFilter && opts.FilterEngine == nil {
	//	return fmt.Errorf(
	//		"proxy %s enable_filter=true but filter engine is nil",
	//		cfg.ID,
	//	)
	//}

	// 5️⃣ 创建 Shadowsocks Server
	server := shadowsocks.NewServer(
		ln,
		c,
		DefaultDirectDialer(),
		func(connID string) traffic.TrafficHook {
			hook := a.newTrafficHook(proxyID, connID, sf)

			return hook
		},
	)

	// 6️⃣ 交给 proxyMgr 管理生命周期
	return a.proxyMgr.StartProxy(proxyID, server)
}

func (a *App) newTrafficHook(proxyID, connID string, sf *SimpleFilter) traffic.TrafficHook {
	return &proxyTrafficHook{
		app:          a,
		proxyID:      proxyID,
		connID:       connID,
		simpleFilter: sf,
	}
}
func (a *App) FilterEngine() *filter.Engine {
	return a.filterEngine
}
func (a *App) SetPluginMgr(p *PluginService) {
	a.pluginMgr = p
}
func (a *App) PluginMgr() *PluginService {
	return a.pluginMgr
}
