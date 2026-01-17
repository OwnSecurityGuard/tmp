package handler

type StartProxyRequest struct {
	BlockIPs   []string `json:"block_ips,omitempty"`
	BlockPorts []string `json:"block_ports,omitempty"`

	PluginName string `json:"plugin_name,omitempty"`
}

type StartProxyResult struct {
	ProxyID string `json:"proxy_id"`

	ListenAddr string `json:"listen_addr"`

	Method   string `json:"method"`
	Password string `json:"password"`

	// 推荐：直接给完整 URL
	ProxyURL string `json:"proxy_url"`

	// 给前端生成二维码用
	QRCodeContent string `json:"qr_code"`
}
