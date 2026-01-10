package filter

import (
	"net"
	"proxy-system-backend/internal/traffic"
)

type CompiledRule struct {
	ID       int64
	Action   Action
	Priority int

	Direction traffic.Direction

	SrcIPNets []*net.IPNet
	DstIPNets []*net.IPNet
	SrcPorts  []PortRange
	DstPorts  []PortRange
}

func (r *CompiledRule) Match(ctx traffic.PacketContext) bool {

	// 1️⃣ 方向匹配
	if r.Direction != traffic.DirectionUnknown &&
		r.Direction != ctx.Direction {
		return false
	}

	// 2️⃣ 源 IP
	if len(r.SrcIPNets) > 0 {
		if ctx.SrcIP == nil || !matchIPNet(ctx.SrcIP, r.SrcIPNets) {
			return false
		}
	}

	// 3️⃣ 目标 IP
	if len(r.DstIPNets) > 0 {
		if ctx.DstIP == nil || !matchIPNet(ctx.DstIP, r.DstIPNets) {
			return false
		}
	}

	// 4️⃣ 源端口
	if len(r.SrcPorts) > 0 {
		if ctx.SrcPort == 0 || !matchPort(ctx.SrcPort, r.SrcPorts) {
			return false
		}
	}

	// 5️⃣ 目标端口
	if len(r.DstPorts) > 0 {
		if ctx.DstPort == 0 || !matchPort(ctx.DstPort, r.DstPorts) {
			return false
		}
	}

	return true
}
func matchIP(addr net.Addr, nets []*net.IPNet) bool {
	tcp, ok := addr.(*net.TCPAddr)
	if !ok {
		return false
	}

	for _, n := range nets {
		if n.Contains(tcp.IP) {
			return true
		}
	}
	return false
}

func matchPort(port int, ranges []PortRange) bool {
	for _, r := range ranges {
		if r.Contains(port) {
			return true
		}
	}
	return false
}

func CompileRule(r Rule) (*CompiledRule, error) {
	cr := &CompiledRule{
		ID:        r.ID,
		Action:    r.Action,
		Priority:  r.Priority,
		Direction: r.Direction,
		SrcPorts:  r.SrcPort,
		DstPorts:  r.DstPort,
	}

	for _, cidr := range r.SrcCIDR {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		cr.SrcIPNets = append(cr.SrcIPNets, n)
	}

	for _, cidr := range r.DstCIDR {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		cr.DstIPNets = append(cr.DstIPNets, n)
	}

	return cr, nil
}
