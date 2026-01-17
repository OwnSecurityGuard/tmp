package app

import (
	"fmt"
	"net"
	"proxy-system-backend/internal/traffic"
	"strconv"
	"strings"
)

type SimpleFilter struct {
	BlockIPs   []*net.IPNet
	BlockPorts []PortRange
}

type PortRange struct {
	From int
	To   int
}

func NewSimpleFilter(blockIPs []string, blockPorts []string) (*SimpleFilter, error) {
	f := &SimpleFilter{}

	for _, ip := range blockIPs {
		_, netw, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, err
		}
		f.BlockIPs = append(f.BlockIPs, netw)
	}

	for _, p := range blockPorts {
		r, err := parsePortRange(p) // "80", "8000-9000"
		if err != nil {
			return nil, err
		}
		f.BlockPorts = append(f.BlockPorts, r)
	}

	return f, nil
}
func (f *SimpleFilter) Match(ctx *traffic.PacketContext) bool {
	ip := ctx.DstIP
	if ip != nil {
		for _, n := range f.BlockIPs {
			if n.Contains(ip) {
				return false
			}
		}
	}

	port := ctx.DstPort
	for _, r := range f.BlockPorts {
		if port >= r.From && port <= r.To {
			return false
		}
	}

	return true
}

func parsePortRange(s string) (PortRange, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return PortRange{}, fmt.Errorf("empty port range")
	}

	// 单端口，例如 "80"
	if !strings.Contains(s, "-") {
		p, err := strconv.Atoi(s)
		if err != nil {
			return PortRange{}, fmt.Errorf("invalid port: %s", s)
		}
		if p < 1 || p > 65535 {
			return PortRange{}, fmt.Errorf("port out of range: %d", p)
		}
		return PortRange{From: p, To: p}, nil
	}

	// 端口范围，例如 "8000-9000"
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return PortRange{}, fmt.Errorf("invalid port range: %s", s)
	}

	from, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return PortRange{}, fmt.Errorf("invalid port range start: %s", s)
	}
	to, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return PortRange{}, fmt.Errorf("invalid port range end: %s", s)
	}

	if from < 1 || from > 65535 || to < 1 || to > 65535 {
		return PortRange{}, fmt.Errorf("port range out of bounds: %s", s)
	}
	if from > to {
		return PortRange{}, fmt.Errorf("invalid port range (from > to): %s", s)
	}

	return PortRange{From: from, To: to}, nil
}
