package shadowsocks

import (
	"errors"
	"io"
	"net"
	"proxy-system-backend/internal/traffic"
)

type proxyConn struct {
	id   string
	hook traffic.TrafficHook
}

func (c *proxyConn) pipe(
	dst net.Conn,
	src net.Conn,
	ctx *traffic.PacketContext,
) error {

	buf := make([]byte, 32*1024)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			ctx.Payload = buf[:n]

			if c.hook != nil && !c.hook.OnPacket(ctx) {
				return errors.New("blocked by hook") // 被过滤，直接断
			}

			if _, werr := dst.Write(ctx.Payload); werr != nil {
				return werr
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
