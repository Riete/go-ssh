package ssh

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type ptyWindowChangeMsg struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

type IShell struct {
	server *Server
	ch     ssh.Channel
	once   sync.Once
	client *ssh.Client
}

type InteractiveShell interface {
	Connect(username string, timeout time.Duration, methods ...AuthMethod) error
	InvokeShell(high, weigh int) error
	ResizePty(high, weigh int) error
	ChanSend(cmd string) error
	ChanRecv(ctx context.Context) chan string
	Close()
}

func (i *IShell) Connect(username string, timeout time.Duration, methods ...AuthMethod) error {
	var err error
	if i.client, err = i.server.connect(username, timeout, methods...); err != nil {
		return errors.New("connect to server failed: " + err.Error())
	}
	return nil
}

func (i *IShell) openChan() error {
	if ch, _, err := i.client.OpenChannel("session", nil); err != nil {
		return errors.New("open channel failed: " + err.Error())
	} else {
		i.ch = ch
		return nil
	}
}

func (i *IShell) setPty(high, weigh int) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	var tm []byte
	for k, v := range modes {
		kv := struct {
			Key byte
			Val uint32
		}{k, v}

		tm = append(tm, ssh.Marshal(&kv)...)
	}
	tm = append(tm, 0)
	req := ptyRequestMsg{
		Term:     "xterm",
		Columns:  uint32(weigh),
		Rows:     uint32(high),
		Width:    uint32(weigh * 8),
		Height:   uint32(high * 8),
		Modelist: string(tm),
	}
	ok, err := i.ch.SendRequest("pty-req", true, ssh.Marshal(&req))
	if err == nil && !ok {
		err = errors.New("ssh: pty-req failed")
	}
	return err
}

func (i *IShell) openShell() error {
	ok, err := i.ch.SendRequest("shell", true, nil)
	if err == nil && !ok {
		return errors.New("ssh: could not start shell")
	}
	return err
}

func (i *IShell) invokeShell(high, weigh int) error {
	if err := i.openChan(); err != nil {
		return err
	}
	if err := i.setPty(high, weigh); err != nil {
		return err
	}
	if err := i.openShell(); err != nil {
		return err
	}
	return nil
}

func (i *IShell) InvokeShell(high, weigh int) error {
	return i.invokeShell(high, weigh)
}

func (i *IShell) ResizePty(high, weigh int) error {
	req := ptyWindowChangeMsg{
		Columns: uint32(weigh),
		Rows:    uint32(high),
		Width:   uint32(weigh * 8),
		Height:  uint32(high * 8),
	}
	_, err := i.ch.SendRequest("window-change", false, ssh.Marshal(&req))
	return err
}

func (i *IShell) ChanSend(cmd string) error {
	_, err := i.ch.Write([]byte(cmd))
	if err == io.EOF {
		i.Close()
		return fmt.Errorf("%s: session closed", err.Error())
	}
	return err
}

func (i *IShell) ChanRecv(ctx context.Context) chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		br := bufio.NewReader(i.ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Millisecond):
				p := make([]byte, 4096)
				count, err := br.Read(p)
				ch <- string(p[0:count])
				if err == io.EOF {
					return
				}
				if err != nil {
					ch <- err.Error()
					return
				}
			}
		}
	}()
	return ch
}

func (i *IShell) Close() {
	i.once.Do(func() {
		_ = i.ch.Close()
		_ = i.client.Close()
	})
}

func NewInteractiveShell(server *Server) InteractiveShell {
	return &IShell{server: server}
}
