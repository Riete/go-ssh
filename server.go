package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	ip   string
	port string
}

func (s Server) connect(username string, timeout time.Duration, methods ...AuthMethod) (*ssh.Client, error) {
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

func NewServer(ip, port string) *Server {
	return &Server{ip: ip, port: port}
}
