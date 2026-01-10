package app

import (
	"context"
	"net"
)

type DirectDialer struct {
	dialer net.Dialer
}

func DefaultDirectDialer() *DirectDialer {
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
