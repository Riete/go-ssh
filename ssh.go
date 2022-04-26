package go_ssh

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshServer struct {
	username  string
	password  string
	ipaddr    string
	port      string
	sshClient *ssh.Client
}

type SSHExecutor interface {
	Cmd(cmd string) error
	CmdGet(cmd string) (string, error)
}

func newSSHServer(username, password, ipaddr, port string) *sshServer {
	return &sshServer{username: username, password: password, ipaddr: ipaddr, port: port}
}

func NewSSHExecutor(username, password, ipaddr, port string) SSHExecutor {
	return newSSHServer(username, password, ipaddr, port)
}

func (s *sshServer) connect() error {
	var err error
	config := &ssh.ClientConfig{
		User: s.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	if s.sshClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:%s", s.ipaddr, s.port), config); err != nil {
		return errors.New(fmt.Sprintf("connect to %s:%s failed, %s", s.ipaddr, s.port, err.Error()))
	}
	return nil
}

func (s *sshServer) openSession() (*ssh.Session, error) {
	if err := s.connect(); err != nil {
		return nil, err
	}
	if session, err := s.sshClient.NewSession(); err != nil {
		return nil, errors.New(fmt.Sprintf("open session failed, %s", err.Error()))
	} else {
		return session, nil
	}
}

func (s *sshServer) Cmd(cmd string) error {
	session, err := s.openSession()
	if err != nil {
		return err
	}
	defer s.sshClient.Close()
	if err := session.Run(cmd); err != nil {
		return errors.New(fmt.Sprintf("run %s failed, %s", cmd, err.Error()))
	}
	return nil
}

func (s *sshServer) CmdGet(cmd string) (string, error) {
	session, err := s.openSession()
	if err != nil {
		return "", err
	}
	defer s.sshClient.Close()
	if output, err := session.CombinedOutput(cmd); err != nil {
		return strings.Trim(string(output), "\n"), errors.New(fmt.Sprintf("run %s failed, %s", cmd, err.Error()))
	} else {
		return strings.Trim(string(output), "\n"), nil
	}
}
