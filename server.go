package go_ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	ip   string
	port string
}

func (s Server) Connect(username string, timeout time.Duration, methods ...AuthMethod) (*ssh.Client, error) {
	config := &ssh.ClientConfig{User: username, HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: timeout}
	for _, method := range methods {
		if auth, err := method(); err != nil {
			return nil, err
		} else {
			config.Auth = append(config.Auth, auth)
		}
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%s", s.ip, s.port), config)
}

func NewServer(ip, port string) *Server {
	return &Server{ip: ip, port: port}
}
