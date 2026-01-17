package main

import (
	"fmt"
	"github.com/hashicorp/go-plugin"
	p "proxy-system-backend/internal/modules/plugin"
	"time"
)

type DemoPlugin struct {
	i int
}

func (dp *DemoPlugin) Decode(payload []byte, isClient bool) (*p.DecodeResult, error) {
	dp.i++
	return &p.DecodeResult{
		IsClient: true,
		Time:     time.Now().Unix(),
		Data:     []byte(fmt.Sprintf(`{ "a": %v}`, payload)),
	}, nil
}

func (dp *DemoPlugin) Encode(data map[string]any) ([]byte, error) {
	return []byte{}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: p.Handshake,
		Plugins: map[string]plugin.Plugin{
			"protocol": &p.ProtocolPluginImpl{
				Impl: &DemoPlugin{},
			},
		},
	})
}
