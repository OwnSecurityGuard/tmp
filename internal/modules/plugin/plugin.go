package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

/**************
 * Client
 **************/

type protocolPluginGRPCClient struct {
	client ProtocolClient
}

func (c *protocolPluginGRPCClient) Decode(payload []byte, isClient bool) (*DecodeResult, error) {
	resp, err := c.client.Decode(context.Background(), &DecodeRequest{
		Payload:  payload,
		IsClient: isClient,
	})
	if err != nil {
		return nil, err
	}

	return &DecodeResult{
		IsClient: resp.IsClient,
		Time:     resp.Time,
		Data:     resp.Data,
	}, nil
}

func (c *protocolPluginGRPCClient) Encode(data []byte) ([]byte, error) {
	resp, err := c.client.Encode(context.Background(), &EncodeRequest{
		Data: data,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

/**************
 * Server
 **************/

type protocolPluginGRPCServer struct {
	UnimplementedProtocolServer
	Impl ProtocolPlugin
}

func (s *protocolPluginGRPCServer) Decode(
	ctx context.Context,
	req *DecodeRequest,
) (*DecodeResponse, error) {
	r, err := s.Impl.Decode(req.Payload, req.IsClient)
	if err != nil {
		return nil, err
	}

	return &DecodeResponse{
		IsClient: r.IsClient,
		Time:     r.Time,
		Data:     r.Data,
	}, nil
}

func (s *protocolPluginGRPCServer) Encode(
	ctx context.Context,
	req *EncodeRequest,
) (*EncodeResponse, error) {
	r, err := s.Impl.Encode(req.Data)
	if err != nil {
		return nil, err
	}

	return &EncodeResponse{Data: r}, nil
}

/**************
 * go-plugin wrapper
 **************/

type ProtocolPluginImpl struct {
	plugin.Plugin
	Impl ProtocolPlugin
}

func (p *ProtocolPluginImpl) GRPCServer(
	broker *plugin.GRPCBroker,
	s *grpc.Server,
) error {
	RegisterProtocolServer(s, &protocolPluginGRPCServer{
		Impl: p.Impl,
	})
	return nil
}

func (p *ProtocolPluginImpl) GRPCClient(
	ctx context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &protocolPluginGRPCClient{
		client: NewProtocolClient(c),
	}, nil
}
