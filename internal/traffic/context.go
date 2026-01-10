package traffic

import (
	"net"
	"time"
)

//
// ===== Direction =====
//

type Direction uint8

const (
	DirectionUnknown Direction = iota // 默认 0 值
	DirectionOut                      // client -> remote
	DirectionIn                       // remote -> client
)

func (d Direction) String() string {
	switch d {
	case DirectionOut:
		return "out"
	case DirectionIn:
		return "in"
	default:
		return "unknown"
	}
}

//
// ===== Protocol（为未来预留，先放好）=====
//

type Protocol uint8

const (
	ProtocolUnknown Protocol = iota
	ProtocolTCP
	ProtocolUDP
)

func (p Protocol) String() string {
	switch p {
	case ProtocolTCP:
		return "tcp"
	case ProtocolUDP:
		return "udp"
	default:
		return "unknown"
	}
}

//
// ===== PacketContext =====
//

type PacketContext struct {
	// 基础标识
	ConnID    string    `json:"conn_id"`
	Direction Direction `json:"direction"`
	Protocol  Protocol  `json:"protocol"`

	// 原始地址（事实来源）
	SrcAddr net.Addr `json:"src_addr"`
	DstAddr net.Addr `json:"dst_addr"`

	// 解析后字段（性能友好）
	SrcIP   net.IP `json:"src_ip"`
	SrcPort int    `json:"src_port"`

	DstIP   net.IP `json:"dst_ip"`
	DstPort int    `json:"dst_port"`

	// 生命周期
	StartAt time.Time `json:"start_at"`

	// 可选：当前 packet payload（filter / plugin 用）
	Payload []byte `json:"payload"`
}

//
// ===== Hook =====
//

type TrafficHook interface {
	// 返回 false 表示中断/丢弃
	OnPacket(ctx *PacketContext) bool
}

//
// ===== Context Factory =====
//

// NewCtx 是唯一推荐的 PacketContext 构造入口
func NewCtx(
	connID string,
	dir Direction,
	proto Protocol,
	src, dst net.Addr,
) *PacketContext {

	ctx := &PacketContext{
		ConnID:    connID,
		Direction: dir,
		Protocol:  proto,
		SrcAddr:   src,
		DstAddr:   dst,
		StartAt:   time.Now(),
	}

	fillIPPort(ctx)
	return ctx
}

// client -> remote
func NewOutCtx(connID string, client, remote net.Conn) *PacketContext {
	return NewCtx(
		connID,
		DirectionOut,
		ProtocolTCP,
		safeRemoteAddr(client),
		safeRemoteAddr(remote),
	)
}

// remote -> client
func NewInCtx(connID string, remote, client net.Conn) *PacketContext {
	return NewCtx(
		connID,
		DirectionIn,
		ProtocolTCP,
		safeRemoteAddr(remote),
		safeRemoteAddr(client),
	)
}

//
// ===== helpers =====
//

// 永远返回 net.Addr，不转 string
func safeRemoteAddr(c net.Conn) net.Addr {
	if c == nil {
		return nil
	}
	return c.RemoteAddr()
}

// 只在 Context 创建时解析一次
func fillIPPort(ctx *PacketContext) {
	if tcp, ok := ctx.SrcAddr.(*net.TCPAddr); ok {
		ctx.SrcIP = tcp.IP
		ctx.SrcPort = tcp.Port
	}

	if tcp, ok := ctx.DstAddr.(*net.TCPAddr); ok {
		ctx.DstIP = tcp.IP
		ctx.DstPort = tcp.Port
	}
}
