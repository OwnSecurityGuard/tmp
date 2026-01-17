package plugin

import (
	"fmt"
	"google.golang.org/grpc"
	"os/exec"
	"sync"
	"time"

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
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*hplugin.Client),
		plugins: make(map[string]ProtocolPlugin),
		infos:   map[string]*PluginInfo{},
	}
}

func (m *Manager) Load(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.infos[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}
	fmt.Println(info.Path)
	client := hplugin.NewClient(&hplugin.ClientConfig{
		GRPCDialOptions: []grpc.DialOption{
			//grpc.WithInsecure(),
		},
		AllowedProtocols: []hplugin.Protocol{
			hplugin.ProtocolGRPC,
		},
		HandshakeConfig: Handshake,
		Plugins: map[string]hplugin.Plugin{
			"protocol": &ProtocolPluginImpl{},
		},
		Cmd: exec.Command(info.Path),
	})

	rpcClient, err := client.Client()
	if err != nil {
		info.Status = PluginStatusError
		info.Error = err.Error()
		return err
	}

	raw, err := rpcClient.Dispense("protocol")
	if err != nil {
		info.Status = PluginStatusError
		info.Error = err.Error()
		return err
	}

	m.clients[name] = client
	m.plugins[name] = raw.(ProtocolPlugin)
	ret, err := m.Decode("test", true, []byte("hhhhhhhhhhhhhhhhhhh"))
	fmt.Println(err)
	fmt.Println("ccc ", ret.Time, ret.Data)
	info.Status = PluginStatusRunning
	info.Error = ""
	info.LoadedAt = time.Now().Unix()
	info.UpdatedAt = info.LoadedAt
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
