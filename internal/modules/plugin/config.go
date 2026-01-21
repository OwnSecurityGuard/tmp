package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config 插件系统配置
type Config struct {
	// 握手配置
	HandshakeConfig HandshakeConfig `json:"handshake_config"`

	// 插件目录配置
	PluginDir string `json:"plugin_dir"`

	// 插件管理配置
	Manager ManagerConfig `json:"manager"`

	// 代理配置
	Proxy ProxyConfig `json:"proxy"`

	// 调试模式配置
	Debug DebugConfig `json:"debug"`
}

// HandshakeConfig 插件握手配置
type HandshakeConfig struct {
	ProtocolVersion  uint   `json:"protocol_version"`
	MagicCookieKey   string `json:"magic_cookie_key"`
	MagicCookieValue string `json:"magic_cookie_value"`
}

// ManagerConfig 插件管理器配置
type ManagerConfig struct {
	// GRPC 配置
	GRPCDialOptions  []string `json:"grpc_dial_options"`
	AllowedProtocols []string `json:"allowed_protocols"`

	// 插件名称映射
	PluginMapping map[string]string `json:"plugin_mapping"`

	// 最大并发插件数量
	MaxConcurrentPlugins int `json:"max_concurrent_plugins"`

	// 插件加载超时时间（秒）
	LoadTimeout int `json:"load_timeout"`

	// 自动加载插件
	AutoLoadPlugins []string `json:"auto_load_plugins"`
}

// DebugConfig 调试配置
type DebugConfig struct {
	// 是否启用调试模式
	Enabled bool `json:"enabled"`

	// 调试模式下的默认插件名
	DefaultPluginName string `json:"default_plugin_name"`

	// 调试模式下的测试数据
	TestData string `json:"test_data"`

	// 是否打印详细日志
	VerboseLogging bool `json:"verbose_logging"`
}

// TrafficHookConfig 流量钩子插件配置
type TrafficHookConfig struct {
	// 流量解码插件名称
	DecoderPlugin string `json:"decoder_plugin"`

	// 是否启用插件解码
	Enabled bool `json:"enabled"`

	// 解码失败时的回退行为（"pass" 通过, "drop" 丢弃, "fallback" 使用备用逻辑）
	FallbackBehavior string `json:"fallback_behavior"`

	// 是否在解码失败时记录日志
	LogDecodeErrors bool `json:"log_decode_errors"`

	// 插件调用超时时间（毫秒）
	TimeoutMs int `json:"timeout_ms"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	// 流量钩子配置
	TrafficHook TrafficHookConfig `json:"traffic_hook"`

	// 自动加载的插件列表（代理启动时）
	AutoLoadPlugins []string `json:"auto_load_plugins"`
}

var (
	configInstance *Config
	configOnce     sync.Once
	configMutex    sync.RWMutex
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		HandshakeConfig: HandshakeConfig{
			ProtocolVersion:  uint(1),
			MagicCookieKey:   "GAME_PROTOCOL_PLUGIN",
			MagicCookieValue: "hello",
		},
		PluginDir: "./data/plugins",
		Manager: ManagerConfig{
			GRPCDialOptions:      []string{},
			AllowedProtocols:     []string{"grpc"},
			PluginMapping:        map[string]string{"protocol": "protocol"},
			MaxConcurrentPlugins: 10,
			LoadTimeout:          30,
			AutoLoadPlugins:      []string{},
		},
		Proxy: ProxyConfig{
			TrafficHook: TrafficHookConfig{
				DecoderPlugin:    "",
				Enabled:          false,
				FallbackBehavior: "pass",
				LogDecodeErrors:  true,
				TimeoutMs:        5000,
			},
			AutoLoadPlugins: []string{},
		},
		Debug: DebugConfig{
			Enabled:           false,
			DefaultPluginName: "",
			TestData:          "",
			VerboseLogging:    false,
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，返回默认配置
		defaultConfig := DefaultConfig()
		configInstance = defaultConfig
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	configInstance = config
	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(configPath string, config *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置（带缩进，便于阅读）
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	configInstance = config
	return nil
}

// GetConfig 获取当前配置实例
func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if configInstance == nil {
		return DefaultConfig()
	}

	return configInstance
}

// UpdateConfig 更新配置
func UpdateConfig(config *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	configInstance = config
}

// IsDebugEnabled 检查是否启用调试模式
func IsDebugEnabled() bool {
	cfg := GetConfig()
	return cfg.Debug.Enabled
}

// IsEnabled IsDebugEnabled 的别名（用于兼容性）
func IsEnabled() bool {
	return IsDebugEnabled()
}

// GetDefaultPluginName 获取调试模式下的默认插件名
func GetDefaultPluginName() string {
	cfg := GetConfig()
	if cfg.Debug.Enabled && cfg.Debug.DefaultPluginName != "" {
		return cfg.Debug.DefaultPluginName
	}
	return ""
}

// GetTestData 获取调试模式下的测试数据
func GetTestData() string {
	cfg := GetConfig()
	if cfg.Debug.Enabled && cfg.Debug.TestData != "" {
		return cfg.Debug.TestData
	}
	return ""
}

// GetTrafficHookDecoderPlugin 获取流量钩子的解码插件名称
func GetTrafficHookDecoderPlugin() string {
	cfg := GetConfig()
	if cfg.Proxy.TrafficHook.DecoderPlugin != "" {
		return cfg.Proxy.TrafficHook.DecoderPlugin
	}
	return ""
}

// IsTrafficHookEnabled 检查流量钩子是否启用
func IsTrafficHookEnabled() bool {
	cfg := GetConfig()
	return cfg.Proxy.TrafficHook.Enabled
}

// GetTrafficHookTimeout 获取流量钩子插件调用超时时间
func GetTrafficHookTimeout() int {
	cfg := GetConfig()
	if cfg.Proxy.TrafficHook.TimeoutMs > 0 {
		return cfg.Proxy.TrafficHook.TimeoutMs
	}
	return 5000 // 默认5秒
}

// ShouldLogDecodeErrors 检查是否记录解码错误
func ShouldLogDecodeErrors() bool {
	cfg := GetConfig()
	return cfg.Proxy.TrafficHook.LogDecodeErrors
}

// GetFallbackBehavior 获取解码失败时的回退行为
func GetFallbackBehavior() string {
	cfg := GetConfig()
	if cfg.Proxy.TrafficHook.FallbackBehavior != "" {
		return cfg.Proxy.TrafficHook.FallbackBehavior
	}
	return "pass" // 默认行为
}
