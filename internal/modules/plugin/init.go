package plugin

import (
	"fmt"
	"log"
	"os"
)

const (
	// DefaultConfigPath 默认配置文件路径
	DefaultConfigPath = "./plugin_config.json"
	// DebugConfigPath 调试模式配置文件路径
	DebugConfigPath = "./plugin_config_debug.json"
)

// InitializePluginSystem 初始化插件系统
// configPath: 配置文件路径，如果为空则使用默认路径
// 返回：Manager实例和错误
func InitializePluginSystem(configPath string) (*Manager, error) {
	// 确定配置文件路径
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// 加载配置
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin config: %w", err)
	}

	// 检查是否应该使用调试模式配置
	debugEnv := os.Getenv("PLUGIN_DEBUG_MODE")
	if debugEnv == "true" || debugEnv == "1" {
		debugConfig, err := LoadConfig(DebugConfigPath)
		if err == nil {
			config = debugConfig
			log.Printf("[Plugin] Debug mode enabled via environment variable")
		} else {
			log.Printf("[Plugin] Warning: failed to load debug config: %v", err)
		}
	}

	// 创建管理器
	manager := NewManager(config)

	// 打印初始化信息
	if config.Debug.VerboseLogging {
		log.Printf("[Plugin] Plugin system initialized")
		log.Printf("[Plugin]   - Plugin dir: %s", config.PluginDir)
		log.Printf("[Plugin]   - Debug mode: %v", config.Debug.Enabled)
		log.Printf("[Plugin]   - Auto-load plugins: %v", config.Manager.AutoLoadPlugins)
		if config.Debug.Enabled {
			log.Printf("[Plugin]   - Default plugin: %s", config.Debug.DefaultPluginName)
		}
	}

	return manager, nil
}

// InitializePluginSystemWithManager 使用现有的管理器初始化插件系统
// manager: 已存在的Manager实例
// configPath: 配置文件路径
func InitializePluginSystemWithManager(manager *Manager, configPath string) error {
	// 确定配置文件路径
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// 加载配置
	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin config: %w", err)
	}

	// 检查是否应该使用调试模式配置
	debugEnv := os.Getenv("PLUGIN_DEBUG_MODE")
	if debugEnv == "true" || debugEnv == "1" {
		debugConfig, err := LoadConfig(DebugConfigPath)
		if err == nil {
			config = debugConfig
			log.Printf("[Plugin] Debug mode enabled via environment variable")
		} else {
			log.Printf("[Plugin] Warning: failed to load debug config: %v", err)
		}
	}

	// 更新管理器配置
	manager.mu.Lock()
	manager.config = config
	manager.mu.Unlock()

	// 打印初始化信息
	if config.Debug.VerboseLogging {
		log.Printf("[Plugin] Manager configuration updated")
	}

	return nil
}
