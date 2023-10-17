package go_ssh

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cloudfoundry/go-socks5"

	"golang.org/x/crypto/ssh"
)

type Socks5Proxy struct {
	client *ssh.Client
	port   string
	logger *log.Logger
}

func (s Socks5Proxy) keepalive() {
	t := time.NewTicker(10 * time.Second)
	for {
		<-t.C
		_, _, err := s.client.SendRequest("keepalive", true, nil)
		if err != nil {
			s.logger.Println("error sending ssh keepalive: " + err.Error())
		}
	}
}

func (s Socks5Proxy) Start() error {
	go s.keepalive()
	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return s.client.Dial(network, addr)
		},
		Logger: s.logger,
	}
	server, err := socks5.New(conf)
	if err != nil {
		return fmt.Errorf("new socks5 server: %s", err)
	}
	return server.ListenAndServe("tcp", "127.0.0.1:"+s.port)
}

func NewSocks5Proxy(client *ssh.Client, port string, logger *log.Logger) *Socks5Proxy {
	if logger == nil {
		logger = log.Default()
	}
	return &Socks5Proxy{client: client, port: port, logger: logger}
}
