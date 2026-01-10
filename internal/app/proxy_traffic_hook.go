package app

import (
	"proxy-system-backend/internal/traffic"
)

type proxyTrafficHook struct {
	proxyID string
	connID  string
	app     *App
}

func (h *proxyTrafficHook) OnPacket(ctx *traffic.PacketContext) bool {
	h.app.Emit(Event{
		Type: EventTraffic,
		Data: map[string]any{
			"proxy_id": h.proxyID,
			"conn_id":  h.connID,
			"payload":  ctx,
		},
	})

	return true
}
