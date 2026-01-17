package app

import (
	"fmt"
	"proxy-system-backend/internal/modules/filter"
	"proxy-system-backend/internal/traffic"
)

type proxyTrafficHook struct {
	proxyID string
	connID  string
	app     *App
	//engine  *filter.Engine todo 后续补充
	simpleFilter *SimpleFilter
}

func (h *proxyTrafficHook) OnPacket(ctx *traffic.PacketContext) bool {
	//engine := h.app.FilterEngine()
	//if engine == nil {
	//	return true
	//}
	//if engine.Match(ctx) {
	//
	//}
	if h.simpleFilter != nil {
		if h.simpleFilter.Match(ctx) {
			fmt.Println("跳过i", ctx.SrcPort, ctx.DstPort)
			return true
		}
	}

	if p := h.app.PluginMgr(); p != nil {
		data, _ := p.Decode("test", true, ctx.Payload)
		h.app.Emit(Event{
			Type: EventParsed,
			Data: data,
		})
		return true
	}

	//h.app.PluginMgr().Load()
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

func (h *proxyTrafficHook) EnableFilter(engine *filter.Engine) {
	h.app.filterEngine = engine
}
