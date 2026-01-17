package plugin

import "encoding/json"

type ProtocolPlugin interface {
	// Decode 对原始 payload 进行解析
	Decode(payload []byte, isClient bool) (*DecodeResult, error)

	Encode(data []byte) ([]byte, error)
}

type DecodeResult struct {
	IsClient bool            `json:"is_client"`
	Time     int64           `json:"time"`
	Data     json.RawMessage `json:"data"`
}
