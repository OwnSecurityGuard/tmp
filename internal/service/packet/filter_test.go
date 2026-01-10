package packet

import (
	"testing"
	"time"

	"proxy-system-backend/internal/types"
)

func TestNewPacketFilter(t *testing.T) {
	pf := NewPacketFilter()

	if pf == nil {
		t.Fatal("NewPacketFilter returned nil")
	}

	// 验证默认值
	config := pf.config.Load().(types.FilterConfig)
	if config.Enabled {
		t.Fatal("Expected filter to be disabled by default")
	}

	if config.DefaultAction != types.FilterActionAllow {
		t.Fatal("Expected default action to be allow")
	}

	if len(config.Rules) != 0 {
		t.Fatal("Expected no rules by default")
	}

	stats := pf.stats.Load().(types.FilterStats)
	if stats.TotalPackets != 0 {
		t.Fatal("Expected total packets to be 0 by default")
	}
}

func TestPacketFilterConfig(t *testing.T) {
	pf := NewPacketFilter()

	// 创建测试配置
	config := types.FilterConfig{
		Enabled:       true,
		DefaultAction: types.FilterActionDeny,
		Rules: []types.FilterRule{
			{
				ID:          1,
				Name:        "Test Rule",
				Action:      types.FilterActionAllow,
				Direction:   types.FilterDirectionIn,
				SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
				SourcePorts: []types.PortRange{{Min: 12345, Max: 12345}}, // 匹配12345端口
				DestIPs:     []types.IPAddress{},                         // 目标IP检查在当前实现中被跳过
				DestPorts:   []types.PortRange{},                         // 目标端口检查实际上检查的是远程地址的端口
				Enabled:     true,
				Priority:    100,
				Description: "测试规则",
			},
		},
		UpdateTime: time.Now(),
	}

	// 测试配置加载
	if err := pf.LoadConfig(config); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 验证配置已加载
	loadedConfig := pf.config.Load().(types.FilterConfig)
	if !loadedConfig.Enabled {
		t.Fatal("Expected filter to be enabled after loading config")
	}

	if loadedConfig.DefaultAction != types.FilterActionDeny {
		t.Fatal("Expected default action to be deny")
	}

	if len(loadedConfig.Rules) != 1 {
		t.Fatal("Expected 1 rule after loading config")
	}

	if loadedConfig.Rules[0].ID != 1 {
		t.Fatal("Expected rule ID to be 1")
	}
}

func TestPacketFilterFilterPacket(t *testing.T) {
	pf := NewPacketFilter()

	// 创建测试配置
	config := types.FilterConfig{
		Enabled:       true,
		DefaultAction: types.FilterActionDeny,
		Rules: []types.FilterRule{
			{
				ID:          1,
				Name:        "Test Rule",
				Action:      types.FilterActionAllow,
				Direction:   types.FilterDirectionIn,
				SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
				SourcePorts: []types.PortRange{{Min: 12345, Max: 12345}},
				DestIPs:     []types.IPAddress{},
				DestPorts:   []types.PortRange{{Min: 8080, Max: 8080}},
				Enabled:     true,
				Priority:    100,
				Description: "测试规则",
			},
		},
		UpdateTime: time.Now(),
	}

	// 加载配置
	if err := pf.LoadConfig(config); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 测试过滤 - 应该允许
	testData := []byte("test data")

	// 添加调试信息
	t.Logf("Testing packet from 192.168.1.10:12345 (should be allowed)")
	result := pf.FilterPacket("client1", "192.168.1.10:12345", testData, types.FilterDirectionIn)
	t.Logf("Filter result: %v", result)

	if !result {
		t.Fatal("Expected packet to be allowed")
	}

	// 测试过滤 - 应该拒绝 (端口不在范围内)
	t.Logf("Testing packet from 192.168.1.10:8080 (should be denied)")
	result = pf.FilterPacket("client1", "192.168.1.10:8080", testData, types.FilterDirectionIn)
	t.Logf("Filter result: %v", result)

	if result {
		t.Fatal("Expected packet to be denied (port not in range)")
	}

	// 测试过滤 - 应该拒绝 (IP不在范围内)
	t.Logf("Testing packet from 10.0.0.10:12345 (should be denied)")
	result = pf.FilterPacket("client1", "10.0.0.10:12345", testData, types.FilterDirectionIn)
	t.Logf("Filter result: %v", result)

	if result {
		t.Fatal("Expected packet to be denied (IP not in range)")
	}
}

func TestPacketFilterDisabled(t *testing.T) {
	pf := NewPacketFilter()

	// 创建禁用过滤的配置
	config := types.FilterConfig{
		Enabled:       false,
		DefaultAction: types.FilterActionDeny,
		Rules: []types.FilterRule{
			{
				ID:          1,
				Name:        "Test Rule",
				Action:      types.FilterActionAllow,
				Direction:   types.FilterDirectionIn,
				SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
				SourcePorts: []types.PortRange{{Min: 1, Max: 60034}},
				DestIPs:     []types.IPAddress{},
				DestPorts:   []types.PortRange{{Min: 8080, Max: 8080}},
				Enabled:     true,
				Priority:    100,
				Description: "测试规则",
			},
		},
		UpdateTime: time.Now(),
	}

	// 加载配置
	if err := pf.LoadConfig(config); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 测试过滤 - 即使配置允许，也应该允许所有数据包
	testData := []byte("test data")
	if !pf.FilterPacket("client1", "10.0.0.10:12345", testData, types.FilterDirectionIn) {
		t.Fatal("Expected packet to be allowed when filter is disabled")
	}
}

func TestPacketFilterNoMatchingRules(t *testing.T) {
	pf := NewPacketFilter()

	// 创建配置 - 没有匹配的规则
	config := types.FilterConfig{
		Enabled:       true,
		DefaultAction: types.FilterActionDeny,
		Rules: []types.FilterRule{
			{
				ID:          1,
				Name:        "Test Rule",
				Action:      types.FilterActionAllow,
				Direction:   types.FilterDirectionOut,
				SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
				SourcePorts: []types.PortRange{{Min: 1, Max: 60034}},
				DestIPs:     []types.IPAddress{},
				DestPorts:   []types.PortRange{{Min: 8080, Max: 8080}},
				Enabled:     true,
				Priority:    100,
				Description: "测试规则",
			},
		},
		UpdateTime: time.Now(),
	}

	// 加载配置
	if err := pf.LoadConfig(config); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 测试过滤 - 没有匹配的规则，使用默认动作
	testData := []byte("test data")
	if pf.FilterPacket("client1", "192.168.1.10:12345", testData, types.FilterDirectionIn) {
		t.Fatal("Expected packet to be denied (no matching rules, default action is deny)")
	}
}

func TestPacketFilterStats(t *testing.T) {
	pf := NewPacketFilter()

	// 创建配置
	config := types.FilterConfig{
		Enabled:       true,
		DefaultAction: types.FilterActionDeny,
		Rules: []types.FilterRule{
			{
				ID:          1,
				Name:        "Allow Rule",
				Action:      types.FilterActionAllow,
				Direction:   types.FilterDirectionIn,
				SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
				SourcePorts: []types.PortRange{{Min: 12345, Max: 12345}}, // 匹配12345端口
				DestIPs:     []types.IPAddress{},                         // 目标IP检查在当前实现中被跳过
				DestPorts:   []types.PortRange{},                         // 目标端口检查实际上检查的是远程地址的端口
				Enabled:     true,
				Priority:    100,
				Description: "测试规则",
			},
		},
		UpdateTime: time.Now(),
	}

	// 加载配置
	if err := pf.LoadConfig(config); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	testData := []byte("test data")

	// 发送多个测试包
	pf.FilterPacket("client1", "192.168.1.10:12345", testData, types.FilterDirectionIn) // 允许
	pf.FilterPacket("client1", "192.168.1.10:12346", testData, types.FilterDirectionIn) // 拒绝
	pf.FilterPacket("client1", "192.168.1.10:12347", testData, types.FilterDirectionIn) // 拒绝

	// 获取统计信息
	stats := pf.GetStats()

	// 验证统计信息
	if stats.TotalPackets != 3 {
		t.Fatalf("Expected total packets to be 3, got %d", stats.TotalPackets)
	}

	if stats.AllowedPackets != 1 {
		t.Fatalf("Expected allowed packets to be 1, got %d", stats.AllowedPackets)
	}

	if stats.BlockedPackets != 2 {
		t.Fatalf("Expected blocked packets to be 2, got %d", stats.BlockedPackets)
	}
}

// 测试GetStats函数
func TestPacketFilterGetStats(t *testing.T) {
	pf := NewPacketFilter()

	// 初始化统计信息
	stats := types.FilterStats{
		TotalPackets:   10,
		AllowedPackets: 8,
		BlockedPackets: 2,
		LastUpdateTime: time.Now(),
	}

	// 直接设置统计信息（跳过LoadConfig中的自动初始化）
	pf.stats.Store(stats)

	// 获取统计信息
	resultStats := pf.GetStats()

	// 验证统计信息
	if resultStats.TotalPackets != 10 {
		t.Fatalf("Expected total packets to be 10, got %d", resultStats.TotalPackets)
	}

	if resultStats.AllowedPackets != 8 {
		t.Fatalf("Expected allowed packets to be 8, got %d", resultStats.AllowedPackets)
	}

	if resultStats.BlockedPackets != 2 {
		t.Fatalf("Expected blocked packets to be 2, got %d", resultStats.BlockedPackets)
	}
}
