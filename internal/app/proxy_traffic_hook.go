package app

import (
	"fmt"
	"proxy-system-backend/internal/modules/filter"
	"proxy-system-backend/internal/modules/plugin"
	"proxy-system-backend/internal/traffic"
	"time"
)

type proxyTrafficHook struct {
	proxyID string
	connID  string
	app     *App
	//engine  *filter.Engine todo 后续补充
	simpleFilter *SimpleFilter
	// 插件调用器
	pluginInvoker *plugin.PluginInvoker
}

func (h *proxyTrafficHook) initPluginInvoker() {
	if h.pluginInvoker == nil {
		// 从 PluginService 获取 Manager
		if mgr := h.app.GetPluginManager(); mgr != nil {
			h.pluginInvoker = plugin.NewPluginInvoker(mgr)
		}
	}
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

	// 使用配置的插件进行解码
	if h.app.PluginMgr() != nil {
		h.initPluginInvoker()

		if plugin.IsTrafficHookEnabled() {
			// 获取配置的解码插件名称
			decoderPlugin := h.getDecoderPlugin()

			if decoderPlugin != "" {
				// 尝试使用配置的插件解码
				data, err := h.decodeWithPlugin(decoderPlugin, ctx)

				if err == nil && data != nil {
					// 解码成功，发送事件
					h.app.Emit(Event{
						Type: EventParsed,
						Data: data,
					})
					return true
				} else {
					// 解码失败，根据回退行为处理
					return h.handleDecodeError(ctx, err, decoderPlugin)
				}
			}
		}
	}

	// 没有配置插件或插件未启用，发送原始流量事件
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

// getDecoderPlugin 获取解码插件名称
func (h *proxyTrafficHook) getDecoderPlugin() string {
	// 优先使用配置的插件
	decoderPlugin := plugin.GetTrafficHookDecoderPlugin()
	if decoderPlugin != "" {
		return decoderPlugin
	}

	// 回退到调试模式的默认插件
	if plugin.IsEnabled() {
		defaultPlugin := plugin.GetDefaultPluginName()
		if defaultPlugin != "" {
			return defaultPlugin
		}
	}

	return ""
}

// decodeWithPlugin 使用插件解码流量数据（使用新的调用器）
func (h *proxyTrafficHook) decodeWithPlugin(pluginName string, ctx *traffic.PacketContext) (*plugin.DecodeResult, error) {
	// 检查调用器是否已初始化
	if h.pluginInvoker == nil {
		h.initPluginInvoker()
	}

	if h.pluginInvoker == nil {
		return nil, fmt.Errorf("plugin invoker not initialized")
	}

	// 将 Direction 转换为 IsClient
	// DirectionOut (client -> remote) = true
	// DirectionIn (remote -> client) = false
	isClient := ctx.Direction == traffic.DirectionOut

	// 创建解码请求
	req := plugin.NewDecodeCallContext(pluginName, isClient, ctx.Payload)

	// 设置上下文参数
	req.SetMetadata("proxy_id", h.proxyID)
	req.SetMetadata("conn_id", h.connID)
	req.SetMetadata("timestamp", time.Now().Unix())
	req.SetMetadata("direction", ctx.Direction.String())
	fmt.Println(req.IsClient)
	// 设置超时
	timeout := time.Duration(plugin.GetTrafficHookTimeout()) * time.Millisecond
	req.Context.SetTimeout(timeout)

	// 启用重试（可选）
	// req.Context.EnableRetryWithConfig(3, 100*time.Millisecond)

	// 调用插件
	return h.pluginInvoker.InvokeDecode(req)
}

// handleDecodeError 处理解码错误
func (h *proxyTrafficHook) handleDecodeError(ctx *traffic.PacketContext, err error, pluginName string) bool {
	if plugin.ShouldLogDecodeErrors() {
		fmt.Printf("[Plugin] Failed to decode with plugin '%s': %v\n", pluginName, err)
	}

	// 根据配置的回退行为处理
	fallbackBehavior := plugin.GetFallbackBehavior()

	switch fallbackBehavior {
	case "drop":
		// 丢弃数据包
		if plugin.ShouldLogDecodeErrors() {
			fmt.Printf("[Plugin] Dropping packet due to fallback behavior 'drop'\n")
		}
		return false

	case "pass":
		// 传递数据包（原始数据）
		h.app.Emit(Event{
			Type: EventTraffic,
			Data: map[string]any{
				"proxy_id":       h.proxyID,
				"conn_id":        h.connID,
				"payload":        ctx,
				"decode_error":   err.Error(),
				"decoder_plugin": pluginName,
			},
		})
		return true

	case "fallback":
		// 使用备用逻辑：尝试使用调试模式的默认插件
		if plugin.IsEnabled() {
			defaultPlugin := plugin.GetDefaultPluginName()
			if defaultPlugin != "" && defaultPlugin != pluginName {
				data, fallbackErr := h.decodeWithPlugin(defaultPlugin, ctx)
				if fallbackErr == nil {
					h.app.Emit(Event{
						Type: EventParsed,
						Data: data,
					})
					return true
				}
			}
		}
		// 最终回退：发送原始流量事件
		h.app.Emit(Event{
			Type: EventTraffic,
			Data: map[string]any{
				"proxy_id":       h.proxyID,
				"conn_id":        h.connID,
				"payload":        ctx,
				"decode_error":   err.Error(),
				"decoder_plugin": pluginName,
			},
		})
		return true

	default:
		// 未知行为，默认为 pass
		h.app.Emit(Event{
			Type: EventTraffic,
			Data: map[string]any{
				"proxy_id":       h.proxyID,
				"conn_id":        h.connID,
				"payload":        ctx,
				"decode_error":   err.Error(),
				"decoder_plugin": pluginName,
			},
		})
		return true
	}
}

func (h *proxyTrafficHook) EnableFilter(engine *filter.Engine) {
	h.app.filterEngine = engine
}
