package types

import (
	"fmt"
	"net"
	"time"
)

// ServerConfig shadowsocks服务器配置
type ServerConfig struct {
	ListenAddr string `json:"listenAddr"`
	Password   string `json:"password"`
	Method     string `json:"method"`
}

// DefaultServerConfig 返回默认配置
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ListenAddr: ":8388",
		Password:   "test-password",
		Method:     "aes-256-gcm",
	}
}

// ServerStatus shadowsocks服务器状态
type ServerStatus struct {
	IsRunning  bool      `json:"isRunning"`
	ListenAddr string    `json:"listenAddr"`
	Password   string    `json:"password"`
	Method     string    `json:"method"`
	Uptime     string    `json:"uptime"`
	Clients    int64     `json:"clients"`
	TotalBytes int64     `json:"totalBytes"`
	StartTime  time.Time `json:"startTime"`
}

// Stats 服务统计信息
type Stats struct {
	TotalRequests int64 `json:"totalRequests"`
	TotalBytes    int64 `json:"totalBytes"`
	ActiveClients int64 `json:"activeClients"`
}

// ServerStats Shadowsocks服务器统计信息
type ServerStats struct {
	IsRunning  bool   `json:"isRunning"`
	ListenAddr string `json:"listenAddr"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Uptime     string `json:"uptime"`
	Clients    int64  `json:"clients"`
	TotalBytes int64  `json:"totalBytes"`
	StartTime  int64  `json:"startTime"`
}

// PacketInfo 数据包信息
type PacketInfo struct {
	Direction  string    `json:"direction"`
	Data       string    `json:"data"` // 十六进制字符串格式
	Length     int       `json:"length"`
	Timestamp  time.Time `json:"timestamp"`
	SourceAddr string    `json:"source_addr"` // 源地址
	DestAddr   string    `json:"dest_addr"`   // 目标地址
}

// ShadowSocksEventCallback ShadowSocks服务器事件回调接口
type ShadowSocksEventCallback interface {
	// OnClientConnected 客户端连接时调用
	OnClientConnected()
	// OnClientDisconnected 客户端断开连接时调用
	OnClientDisconnected()
	// OnPacketReceived 收到数据包时调用
	OnPacketReceived(direction string, data []byte, length int, sourceAddr, destAddr string)
	// OnProxyConnectionStarted 代理连接开始时调用
	OnProxyConnectionStarted(clientID, connectionID string)
	// OnProxyConnectionEnded 代理连接结束时调用
	OnProxyConnectionEnded(clientID string)
}

// FilterAction 过滤动作类型
type FilterAction string

const (
	FilterActionAllow FilterAction = "allow" // 允许
	FilterActionDeny  FilterAction = "deny"  // 拒绝
)

// FilterDirection 过滤方向类型
type FilterDirection string

const (
	FilterDirectionIn   FilterDirection = "in"   // 上行（客户端到服务器）
	FilterDirectionOut  FilterDirection = "out"  // 下行（服务器到客户端）
	FilterDirectionBoth FilterDirection = "both" // 双向
)

// IPAddress IP地址结构
type IPAddress struct {
	IP   string `json:"ip"`   // IP地址
	Mask int    `json:"mask"` // 子网掩码位数
}

func (a IPAddress) CIDR() (string, error) {
	ip := net.ParseIP(a.IP)
	if ip == nil {
		return "", fmt.Errorf("invalid ip: %s", a.IP)
	}

	// 判断 IPv4 / IPv6
	var maxMask int
	if ip.To4() != nil {
		maxMask = 32
	} else {
		maxMask = 128
	}

	if a.Mask < 0 || a.Mask > maxMask {
		return "", fmt.Errorf("invalid mask %d for ip %s", a.Mask, a.IP)
	}

	return fmt.Sprintf("%s/%d", ip.String(), a.Mask), nil
}

// PortRange 端口范围结构

// FilterRule 过滤规则结构
type FilterRule struct {
	ID        int             `json:"id"`         // 规则ID
	Name      string          `json:"name"`       // 规则名称
	Action    FilterAction    `json:"action"`     // 动作（允许/拒绝）
	Direction FilterDirection `json:"direction"`  // 方向
	SourceIPs []IPAddress     `json:"source_ips"` // 源IP列表（空列表表示任意）
	//SourcePorts []PortRange     `json:"source_ports"` // 源端口列表（空列表表示任意）
	DestIPs []IPAddress `json:"dest_ips"` // 目标IP列表（空列表表示任意）
	//DestPorts   []PortRange     `json:"dest_ports"`   // 目标端口列表（空列表表示任意）
	Enabled     bool      `json:"enabled"`     // 是否启用
	Priority    int       `json:"priority"`    // 优先级（数字越大优先级越高）
	Description string    `json:"description"` // 规则描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FilterConfig 过滤配置
type FilterConfig struct {
	Enabled       bool         `json:"enabled"`        // 是否启用过滤
	DefaultAction FilterAction `json:"default_action"` // 默认动作
	Rules         []FilterRule `json:"rules"`          // 规则列表
	UpdateTime    time.Time    `json:"update_time"`    // 最后更新时间
}

// FilterStats 过滤统计信息
type FilterStats struct {
	TotalPackets   int64     `json:"total_packets"`    // 总包数
	AllowedPackets int64     `json:"allowed_packets"`  // 允许的包数
	BlockedPackets int64     `json:"blocked_packets"`  // 阻止的包数
	ActiveRules    int       `json:"active_rules"`     // 活跃规则数
	LastUpdateTime time.Time `json:"last_update_time"` // 最后更新时间
}
