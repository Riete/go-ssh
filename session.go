package go_ssh

import (
	"errors"
	"time"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	server *Server
	client *ssh.Client
}

type SessionExecutor interface {
	Connect(username string, timeout time.Duration, methods ...AuthMethod) error
	Cmd(cmd string) error
	CmdGet(cmd string) ([]byte, error)
	RawClient() *ssh.Client
	Close()
}

func (s *Session) Connect(username string, timeout time.Duration, methods ...AuthMethod) error {
	var err error
	if s.client, err = s.server.Connect(username, timeout, methods...); err != nil {
		return errors.New("connect to server failed: " + err.Error())
	}
	return nil
}

func (s *Session) openSession() (*ssh.Session, error) {
	if session, err := s.client.NewSession(); err != nil {
		return nil, errors.New("open session failed: " + err.Error())
	} else {
		return session, nil
	}
}

func (s *Session) Cmd(cmd string) error {
	if session, err := s.openSession(); err != nil {
		return err
	} else {
		return session.Run(cmd)
	}
}

func (s *Session) CmdGet(cmd string) ([]byte, error) {
	if session, err := s.openSession(); err != nil {
		return nil, err
	} else {
		return session.CombinedOutput(cmd)
	}
}

func (s *Session) RawClient() *ssh.Client {
	return s.client
}

func (s *Session) Close() {
	_ = s.client.Close()
}

func NewSessionExecutor(server *Server) SessionExecutor {
	return &Session{server: server}
}
