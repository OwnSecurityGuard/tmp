package main

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"proxy-system-backend/cmd/test/common"
	plugin2 "proxy-system-backend/proto/plugin"
	"proxy-system-backend/proto/visual"
)

type DemoPlugin struct{}

func (d *DemoPlugin) Info(ctx context.Context) (*plugin2.PluginInfo, error) {
	return &plugin2.PluginInfo{
		Name:        "demo-plugin",
		Game:        "demo-game",
		Version:     "v1.0.0",
		Author:      "you",
		Description: "demo protocol plugin",
		Protocols: []*plugin2.ProtocolDesc{
			{Id: "v1", DisplayName: "Demo Protocol v1"},
		},
	}, nil
}

func (d *DemoPlugin) Decode(
	ctx context.Context,
	req *plugin2.DecodeRequest,
) (*plugin2.DecodeResponse, error) {

	// 假装解码
	msg := &visual.VisualMessage{
		Id:        "12",
		ConnId:    "213",
		Timestamp: 23,
		Direction: 0,
		Name:      "",
		Category:  "",
		Summary:   "",
		Fields:    nil,
		Raw:       nil,
		Tags:      nil,
	}

	return &plugin2.DecodeResponse{
		Messages: []*visual.VisualMessage{msg},
	}, nil
}

func (d *DemoPlugin) ResetConn(ctx context.Context, req *plugin2.ResetConnRequest) error {
	// demo 什么都不做
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "GAME_PROTOCOL_PLUGIN",
			MagicCookieValue: "hello",
		},
		Plugins: map[string]plugin.Plugin{
			"game_protocol": &common.GRPCPlugin{
				Impl: &DemoPlugin{},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
