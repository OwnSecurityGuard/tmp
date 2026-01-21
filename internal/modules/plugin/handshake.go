package plugin

import (
	"github.com/hashicorp/go-plugin"
)

// GetHandshakeConfig 获取握手配置（从配置文件读取）
func GetHandshakeConfig() plugin.HandshakeConfig {
	cfg := GetConfig()
	return plugin.HandshakeConfig{
		ProtocolVersion:  cfg.HandshakeConfig.ProtocolVersion,
		MagicCookieKey:   cfg.HandshakeConfig.MagicCookieKey,
		MagicCookieValue: cfg.HandshakeConfig.MagicCookieValue,
	}
}

// Handshake 已弃用：保留用于向后兼容
// 请使用 GetHandshakeConfig() 代替
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GAME_PROTOCOL_PLUGIN",
	MagicCookieValue: "hello",
}
