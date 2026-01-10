package app

type EventType string

const (
	EventProxyStarted EventType = "proxy_started"
	EventProxyStopped EventType = "proxy_stopped"
	EventRuleUpdated  EventType = "rule_updated"
	EventPluginLoaded EventType = "plugin_loaded"
	EventTraffic      EventType = "EventTraffic"
)

type Event struct {
	Type EventType
	Data any
}
