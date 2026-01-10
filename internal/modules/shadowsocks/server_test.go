package shadowsocks

import (
	"context"
	"fmt"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"net"
	"proxy-system-backend/internal/traffic"
	"testing"
)

type TestHook struct {
}
type DirectDialer struct {
	dialer net.Dialer
}

func NewDirectDialer() *DirectDialer {
	return &DirectDialer{
		dialer: net.Dialer{},
	}
}

func (d *DirectDialer) DialContext(
	ctx context.Context,
	network, addr string,
) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, addr)
}
func (t TestHook) OnPacket(ctx *traffic.PacketContext) bool {
	//TODO implement me
	fmt.Println(ctx.SrcAddr, ctx.DstAddr, ctx.DstAddr, ctx.DstAddr)
	return true
}

func TestNewServer(t *testing.T) {

	cipher, _ := core.PickCipher("aes-256-gcm", nil, "test-password")
	ln, _ := net.Listen("tcp", ":8388")
	s := NewServer(ln, NewDirectDialer(), func(connID string) traffic.TrafficHook {

	})
	s.cipher = cipher

	for {
		c, _ := ln.Accept()
		go s.handleConn(c)
	}
}
