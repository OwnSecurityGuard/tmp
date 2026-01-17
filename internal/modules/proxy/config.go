package proxy

import (
	"fmt"
	"github.com/shadowsocks/go-shadowsocks2/core"
)

type Config struct {
	// ===== 身份 =====
	ID   string `json:"id"`   // proxy_id（稳定标识）
	Name string `json:"name"` // 显示用

	// ===== 网络 =====
	ListenAddr string `json:"listen_addr"` // ":8388"

	// ===== Shadowsocks =====
	Method   string `json:"method"` // aes-256-gcm
	Password string `json:"password"`

	// ===== 行为配置 =====
	EnableFilter bool `json:"enable_filter"`

	// ===== 生命周期 =====
	Enabled   bool  `json:"enabled"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`

	BlockIPs   []string `json:"block_ips,omitempty"`
	BlockPorts []string `json:"block_ports,omitempty"`
}

func (c *Config) BuildCipher() (core.Cipher, error) {
	if c.Method == "" {
		return nil, fmt.Errorf("cipher method is empty")
	}
	if c.Password == "" {
		return nil, fmt.Errorf("cipher password is empty")
	}

	cipher, err := core.PickCipher(
		c.Method,
		nil, // key is derived from password
		c.Password,
	)
	if err != nil {
		return nil, fmt.Errorf("pick cipher failed: %w", err)
	}

	return cipher, nil
}
