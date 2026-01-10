package shadowsocks

import (
	"context"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"net"
	"proxy-system-backend/internal/modules/shared"
	"proxy-system-backend/internal/traffic"
	"sync"
)

type Server struct {
	listener net.Listener
	hookFn   func(connID string) traffic.TrafficHook
	wg       sync.WaitGroup
	dialer   Dialer
	cipher   core.Cipher

	closeOnce sync.Once
	closed    chan struct{}
}

func NewServer(l net.Listener, c core.Cipher, d Dialer, hf func(connID string) traffic.TrafficHook) *Server {
	return &Server{
		listener: l,
		dialer:   d,
		hookFn:   hf,
		cipher:   c,
		closed:   make(chan struct{}),
	}
}
func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		//fmt.Println("debug new conn")
		if err != nil {
			select {
			case <-s.closed:
				return nil // 正常关闭
			default:
				return err // 异常错误
			}
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handleConn(c)
		}(conn)
	}
}

func (s *Server) handleConn(client net.Conn) {
	defer client.Close()

	// 1️⃣ Shadowsocks 解密
	ssConn := s.unwrap(client)
	defer ssConn.Close()

	// 2️⃣ 读取目标地址（SOCKS 协议）
	target, err := socks.ReadAddr(ssConn)
	if err != nil {
		return
	}

	// 3️⃣ 连接目标
	remote, err := s.dialer.DialContext(context.Background(), "tcp", target.String())
	if err != nil {
		return
	}
	defer remote.Close()

	// 4️⃣ 双向 pipe（你现在已有的）
	connID := shared.GenerateConnID()
	hook := s.hookFn(connID)
	pc := &proxyConn{id: connID, hook: hook}

	errCh := make(chan error, 2)

	go func() {
		errCh <- pc.pipe(remote, ssConn, traffic.NewOutCtx(connID, client, remote))
	}()

	go func() {
		errCh <- pc.pipe(ssConn, remote, traffic.NewInCtx(connID, remote, client))
	}()

	<-errCh
}

func (s *Server) unwrap(conn net.Conn) net.Conn {
	if s.cipher == nil {
		return conn
	}
	return s.cipher.StreamConn(conn)
}
func (s *Server) Close() error {
	var err error

	s.closeOnce.Do(func() {
		close(s.closed)

		if s.listener != nil {
			err = s.listener.Close()
		}

		// 等待所有 handleConn 完成
		s.wg.Wait()
	})

	return err
}
