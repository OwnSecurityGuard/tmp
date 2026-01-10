package filter

import (
	"net"
	"proxy-system-backend/internal/traffic"
)

type Action string

const (
	ActionAllow Action = "allow"
	ActionDeny  Action = "deny"
)

type Config struct {
	Enabled       bool
	DefaultAction Action
	Rules         []Rule
}

type Rule struct {
	ID          int64
	Name        string
	Description string

	Action    Action
	Direction traffic.Direction
	Priority  int
	Enabled   bool

	// ===== 匹配条件（未编译）=====
	SrcCIDR []string
	DstCIDR []string
	SrcPort []PortRange
	DstPort []PortRange

	Tags []string
}

func matchIPNet(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

type PortRange struct {
	Min int `json:"min"` // 最小端口
	Max int `json:"max"` // 最大端口
}

func (r PortRange) Contains(p int) bool {
	return p >= r.Min && p <= r.Max
}
