package handler

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"proxy-system-backend/internal/app"
	"proxy-system-backend/internal/modules/proxy"
	"proxy-system-backend/internal/modules/shared"
	"time"
)

type ProxyHandler struct {
	app *app.App
}

func NewProxyHandler(app *app.App) *ProxyHandler {
	return &ProxyHandler{app: app}
}

func (h *ProxyHandler) StartProxy(c *gin.Context) {
	var req StartProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 补生命周期字段
	n, _ := allocateRandomPort()
	ip := "192.168.10.5" //getLocalLANIP()
	now := time.Now().Unix()
	cfg := proxy.Config{}
	fmt.Println(ip, n)
	cfg.Password = "test-password" //shared.GenerateConnID()
	cfg.ID = shared.GenerateConnID()
	cfg.Method = "aes-256-gcm"
	cfg.Name = "tmp"
	cfg.Enabled = true
	cfg.CreatedAt = now
	cfg.UpdatedAt = now
	cfg.BlockIPs = req.BlockIPs
	cfg.BlockPorts = req.BlockPorts
	cfg.ListenAddr = fmt.Sprintf("%s:%v", ip, n)
	fmt.Println(fmt.Sprintf("%+v", cfg))
	if err := h.app.StartProxy(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	dat, _ := BuildSSQRCodeContent(cfg)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		//"result": req.ID,
		"qr_code": dat,
	})
}

func allocateRandomPort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func getLocalLANIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// 跳过 down 的、loopback 的
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // 跳过 IPv6
			}

			if isPrivateIPv4(ip) {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no private LAN ip found")
}
func isPrivateIPv4(ip net.IP) bool {
	if ip.IsLoopback() {
		return false
	}

	switch {
	case ip[0] == 10:
		return true
	case ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31:
		return true
	case ip[0] == 192 && ip[1] == 168:
		return true
	default:
		return false
	}
}

func BuildSSQRCodeContent(cfg proxy.Config) (string, error) {
	if cfg.Method == "" || cfg.Password == "" {
		return "", fmt.Errorf("method or password is empty")
	}
	if cfg.ListenAddr == "" {
		return "", fmt.Errorf("listen_addr is empty")
	}

	// 解析端口
	host, portStr, err := net.SplitHostPort(cfg.ListenAddr)
	if err != nil {
		return "", fmt.Errorf("invalid listen_addr: %w", err)
	}

	// method:password@host:port
	raw := fmt.Sprintf(
		"%s:%s@%s:%s",
		cfg.Method,
		cfg.Password,
		host,
		portStr,
	)

	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	return "ss://" + encoded, nil
}
