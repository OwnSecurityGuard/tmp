package plugin

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"google.golang.org/grpc"

	hplugin "github.com/hashicorp/go-plugin"
)

type PluginStatus string

const (
	PluginStatusRegistered PluginStatus = "registered" // 已注册，未启动
	PluginStatusRunning    PluginStatus = "running"    // 已加载
	PluginStatusStopped    PluginStatus = "stopped"    // 已卸载
	PluginStatusError      PluginStatus = "error"
)

type PluginInfo struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Status    PluginStatus `json:"status"`
	Error     string       `json:"error,omitempty"`
	LoadedAt  int64        `json:"loaded_at,omitempty"`
	UpdatedAt int64        `json:"updated_at"`
}

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*hplugin.Client
	plugins map[string]ProtocolPlugin
	infos   map[string]*PluginInfo
	config  *Config // 动态配置
}

func NewManager(config *Config) *Manager {
	if config == nil {
		config = GetConfig()
	}
	return &Manager{
		clients: make(map[string]*hplugin.Client),
		plugins: make(map[string]ProtocolPlugin),
		infos:   map[string]*PluginInfo{},
		config:  config,
	}
}

func (m *Manager) Load(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.infos[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	if m.config.Debug.VerboseLogging {
		fmt.Printf("[Plugin] Loading plugin: %s from %s\n", name, info.Path)
	}

	// 获取握手配置
	handshakeConfig := GetHandshakeConfig()

	// 构建 GRPC 拨号选项
	var grpcDialOptions []grpc.DialOption
	for _, optName := range m.config.Manager.GRPCDialOptions {
		// 根据配置添加 GRPC 拨号选项
		// 这里可以根据需要扩展支持更多选项
		switch optName {
		case "insecure":
			// grpcDialOptions = append(grpcDialOptions, grpc.WithInsecure())
		default:
			// 未知选项，忽略或记录警告
		}
	}

	// 确定允许的协议
	var allowedProtocols []hplugin.Protocol
	for _, protoName := range m.config.Manager.AllowedProtocols {
		switch protoName {
		case "grpc":
			allowedProtocols = append(allowedProtocols, hplugin.ProtocolGRPC)
		case "netrpc":
			allowedProtocols = append(allowedProtocols, hplugin.ProtocolNetRPC)
		}
	}

	// 构建插件映射
	plugins := map[string]hplugin.Plugin{}
	for internalName, pluginType := range m.config.Manager.PluginMapping {
		switch pluginType {
		case "protocol":
			plugins[internalName] = &ProtocolPluginImpl{}
		}
	}

	// 创建插件客户端
	client := hplugin.NewClient(&hplugin.ClientConfig{
		GRPCDialOptions:  grpcDialOptions,
		AllowedProtocols: allowedProtocols,
		HandshakeConfig:  handshakeConfig,
		Plugins:          plugins,
		Cmd:              exec.Command(info.Path),
	})

	rpcClient, err := client.Client()
	if err != nil {
		info.Status = PluginStatusError
		info.Error = err.Error()
		return fmt.Errorf("failed to get RPC client: %w", err)
	}

	raw, err := rpcClient.Dispense("protocol")
	if err != nil {
		info.Status = PluginStatusError
		info.Error = err.Error()
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	m.clients[name] = client
	m.plugins[name] = raw.(ProtocolPlugin)

	// 调试模式下执行测试调用（仅在启用调试模式且配置了测试数据时）
	if IsDebugEnabled() {
		testData := GetTestData()
		if testData != "" {
			ret, err := m.Decode(name, true, []byte(testData))
			if m.config.Debug.VerboseLogging {
				if err != nil {
					fmt.Printf("[Plugin Debug] Test decode error: %v\n", err)
				} else {
					fmt.Printf("[Plugin Debug] Test decode result: time=%d, data=%s\n", ret.Time, ret.Data)
				}
			}
		}
	}

	info.Status = PluginStatusRunning
	info.Error = ""
	info.LoadedAt = time.Now().Unix()
	info.UpdatedAt = info.LoadedAt

	if m.config.Debug.VerboseLogging {
		fmt.Printf("[Plugin] Plugin %s loaded successfully\n", name)
	}

	return nil
}

func (m *Manager) Decode(
	name string,
	isClient bool,
	payload []byte,
) (*DecodeResult, error) {
	p, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not loaded", name)
	}
	return p.Decode(payload, isClient)
}

func (m *Manager) Unload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[name]
	if !ok {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	client.Kill()
	delete(m.clients, name)
	delete(m.plugins, name)

	info := m.infos[name]
	info.Status = PluginStatusStopped
	info.UpdatedAt = time.Now().Unix()
	return nil
}

func (m *Manager) List() []*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*PluginInfo
	for _, info := range m.infos {
		list = append(list, info)
	}
	return list
}

func (m *Manager) Get(name string) (*PluginInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.infos[name]
	return info, ok
}

func (m *Manager) Register(name, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.infos[name]; ok {
		return fmt.Errorf("plugin %s already exists", name)
	}

	m.infos[name] = &PluginInfo{
		Name:      name,
		Path:      path,
		Status:    PluginStatusRegistered,
		UpdatedAt: time.Now().Unix(),
	}
	return nil
}
