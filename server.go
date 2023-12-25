package ssh

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/proxy"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	ip     string
	port   string
	dialer func(ctx context.Context, network, addr string) (c net.Conn, err error)
}

func (s *Server) connect(username string, timeout time.Duration, methods ...AuthMethod) (*ssh.Client, error) {
	config := &ssh.ClientConfig{User: username, HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: timeout}
	for _, method := range methods {
		if auth, err := method(); err != nil {
			return nil, err
		} else {
			config.Auth = append(config.Auth, auth)
		}
	}
	addr := s.ip + ":" + s.port
	if s.dialer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := s.dialer(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
		if err != nil {
			return nil, err
		}
		return ssh.NewClient(c, chans, reqs), nil
	}
	return ssh.Dial("tcp", addr, config)
}

func (s *Server) OpenSession() SessionExecutor {
	return NewSessionExecutor(s)
}

func (s *Server) OpenSftp() SftpExecutor {
	return NewSftpExecutor(s)
}

func (s *Server) OpenIShell() InteractiveShell {
	return NewInteractiveShell(s)
}

func (s *Server) OpenClient(username string, timeout time.Duration, methods ...AuthMethod) (*ssh.Client, error) {
	return s.connect(username, timeout, methods...)
}

func (s *Server) SetSocks5Proxy(addr string, auth *proxy.Auth) error {
	d, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return err
	}
	xd := d.(proxy.ContextDialer)
	s.dialer = xd.DialContext
	return nil
}

func NewServer(ip, port string) *Server {
	return &Server{ip: ip, port: port}
}
