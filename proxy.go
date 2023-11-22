package ssh

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/armon/go-socks5"

	"golang.org/x/crypto/ssh"
)

type Socks5Proxy struct {
	client *ssh.Client
	port   string
}

func (s Socks5Proxy) keepalive() {
	t := time.NewTicker(10 * time.Second)
	for {
		<-t.C
		s.client.SendRequest("keepalive", true, nil)
	}
}

func (s Socks5Proxy) ListenAndServe() error {
	go s.keepalive()
	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return s.client.Dial(network, addr)
		},
	}
	server, err := socks5.New(conf)
	if err != nil {
		return fmt.Errorf("new socks5 server: %s", err)
	}
	return server.ListenAndServe("tcp", "127.0.0.1:"+s.port)
}

func NewSocks5Proxy(client *ssh.Client, port string) *Socks5Proxy {
	return &Socks5Proxy{client: client, port: port}
}
