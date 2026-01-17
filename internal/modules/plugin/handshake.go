package plugin

import "github.com/hashicorp/go-plugin"

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GAME_PROTOCOL_PLUGIN",
	MagicCookieValue: "hello",
}
