package go_ssh

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func (h *Host) connect() error {
	config := &ssh.ClientConfig{
		User: h.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(h.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	if client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", h.IpAddr, h.Port), config); err != nil {
		return errors.New(fmt.Sprintf("connect to %s:%s failed, %s", h.IpAddr, h.Port, err.Error()))
	} else {
		h.SSHClient = client
	}
	return nil
}

func (h *Host) openSession() (*ssh.Session, error) {
	if err := h.connect(); err != nil {
		return nil, err
	}
	if session, err := h.SSHClient.NewSession(); err != nil {
		return nil, errors.New(fmt.Sprintf("open session failed, %s", err.Error()))
	} else {
		return session, nil
	}
}

func (h *Host) Cmd(cmd string) error {
	session, err := h.openSession()
	if err != nil {
		return err
	}
	defer h.SSHClient.Close()
	if err := session.Run(cmd); err != nil {
		return errors.New(fmt.Sprintf("run %s failed, %s", cmd, err.Error()))
	}
	return nil
}

func (h *Host) CmdGet(cmd string) (string, error) {
	session, err := h.openSession()
	if err != nil {
		return "", err
	}
	defer h.SSHClient.Close()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stderr = &stderr
	session.Stdout = &stdout
	if err := session.Run(cmd); err != nil {
		return strings.Trim(stderr.String(), "\n"), errors.New(fmt.Sprintf("run %s failed, %s", cmd, err.Error()))
	} else {
		return strings.Trim(stdout.String(), "\n"), nil
	}
}
