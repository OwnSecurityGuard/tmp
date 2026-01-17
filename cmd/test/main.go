package main

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-plugin"
	p "proxy-system-backend/internal/modules/plugin"
)

type DemoPlugin struct{}

func (d *DemoPlugin) Decode(payload []byte, isClient bool) (*p.DecodeResult, error) {
	return &p.DecodeResult{
		IsClient: isClient,
		Time:     time.Now().UnixMilli(),
		Data:     []byte(fmt.Sprintf(`{"raw":%q}`, payload)),
	}, nil
}

func (d *DemoPlugin) Encode(data []byte) ([]byte, error) {
	return data, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: p.Handshake,
		Plugins: map[string]plugin.Plugin{
			"protocol": &p.ProtocolPluginImpl{
				Impl: &DemoPlugin{},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
