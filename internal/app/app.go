package app

import (
	"fmt"
	"net"
	"proxy-system-backend/internal/modules/filter"
	"proxy-system-backend/internal/modules/plugin"
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

	// 3️⃣ 自动加载配置的插件
	if a.pluginMgr != nil {
		// 获取需要自动加载的插件列表
		pluginConfig := plugin.GetConfig()
		for _, pluginName := range pluginConfig.Manager.AutoLoadPlugins {
			if err := a.pluginMgr.Load(pluginName); err != nil {
				// 如果插件加载失败，记录警告但继续运行
				fmt.Printf("[Warning] Failed to load plugin '%s': %v\n", pluginName, err)
			} else {
				if pluginConfig.Debug.VerboseLogging {
					fmt.Printf("[Plugin] Auto-loaded plugin: %s\n", pluginName)
				}
			}
		}

		// 调试模式下加载默认插件
		if plugin.IsEnabled() {
			defaultPluginName := plugin.GetDefaultPluginName()
			if defaultPluginName != "" {
				if err := a.pluginMgr.Load(defaultPluginName); err != nil {
					fmt.Printf("[Warning] Debug mode: Failed to load default plugin '%s': %v\n", defaultPluginName, err)
				} else {
					fmt.Printf("[Plugin Debug] Loaded default plugin: %s\n", defaultPluginName)
				}
			}
		}
	}

	// 4️⃣ 构建加密器
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
func (a *App) GetPluginManager() *plugin.Manager {
	if a.pluginMgr != nil {
		return a.pluginMgr.GetPluginManager()
	}
	return nil
}
