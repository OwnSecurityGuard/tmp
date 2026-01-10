package common

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	plugin2 "proxy-system-backend/proto/plugin"
)

var PluginMap = map[string]plugin.Plugin{
	"kv_grpc": &GRPCPlugin{},
}

type GameProtocol interface {
	Info(ctx context.Context) (*plugin2.PluginInfo, error)
	Decode(ctx context.Context, req *plugin2.DecodeRequest) (*plugin2.DecodeResponse, error)
	ResetConn(ctx context.Context, req *plugin2.ResetConnRequest) error
}

type GRPCPlugin struct {
	plugin.Plugin
	Impl GameProtocol
}

func (p *GRPCPlugin) GRPCServer(
	broker *plugin.GRPCBroker,
	s *grpc.Server,
) error {
	plugin2.RegisterGameProtocolPluginServer(
		s,
		&GRPCServer{Impl: p.Impl},
	)
	return nil
}

func (p *GRPCPlugin) GRPCClient(
	ctx context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &GRPCClient{
		client: plugin2.NewGameProtocolPluginClient(c),
	}, nil
}

type GRPCClient struct {
	client plugin2.GameProtocolPluginClient
}

func (c *GRPCClient) Info(ctx context.Context) (*plugin2.PluginInfo, error) {
	return c.client.Info(ctx, &emptypb.Empty{})
}

func (c *GRPCClient) Decode(
	ctx context.Context,
	req *plugin2.DecodeRequest,
) (*plugin2.DecodeResponse, error) {
	return c.client.Decode(ctx, req)
}

func (c *GRPCClient) ResetConn(
	ctx context.Context,
	req *plugin2.ResetConnRequest,
) error {
	_, err := c.client.ResetConn(ctx, req)
	return err
}

type GRPCServer struct {
	plugin2.UnimplementedGameProtocolPluginServer
	Impl GameProtocol
}

func (s *GRPCServer) Info(ctx context.Context, _ *emptypb.Empty) (*plugin2.PluginInfo, error) {
	return s.Impl.Info(ctx)
}

func (s *GRPCServer) Decode(ctx context.Context, req *plugin2.DecodeRequest) (*plugin2.DecodeResponse, error) {
	return s.Impl.Decode(ctx, req)
}

func (s *GRPCServer) ResetConn(ctx context.Context, req *plugin2.ResetConnRequest) (*emptypb.Empty, error) {
	err := s.Impl.ResetConn(ctx, req)
	return &emptypb.Empty{}, err
}
