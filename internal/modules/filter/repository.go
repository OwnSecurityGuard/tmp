package filter

import "context"

type Repository interface {
	LoadConfig(ctx context.Context) ([]*Config, error)
	// SaveConfig 保存完整配置（管理端使用）
	SaveConfig(ctx context.Context, cfg *Config) error
}
