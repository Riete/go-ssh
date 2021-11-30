package go_ssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

type iShell struct {
	*sshServer
	ch ssh.Channel
}

type InteractiveShell interface {
	InvokeShell(high, weigh int) error
	ResizePty(high, weigh int) error
	ChanSend(cmd string) error
	ChanRcv(ch chan string)
	ChanClose()
}

func NewInteractiveShell(username, password, ipaddr, port string) InteractiveShell {
	return &iShell{sshServer: newSSHServer(username, password, ipaddr, port)}
}

func (i *iShell) openChan() error {
	if err := i.connect(); err != nil {
		return err
	}
	ch, _, err := i.sshClient.OpenChannel("session", nil)
	if err != nil {
		return errors.New(fmt.Sprintf("open channel failed, %s", err.Error()))
	}
	i.ch = ch
	return nil
}

func (i iShell) setPty(high, weigh int) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
		ssh.VSTATUS:       1,
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

func (i iShell) openShell() error {
	ok, err := i.ch.SendRequest("shell", true, nil)
	if err == nil && !ok {
		return errors.New("ssh: could not start shell")
	}
	return err
}

func (i *iShell) invokeShell(high, weigh int) error {
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

func (i *iShell) InvokeShell(high, weigh int) error {
	return i.invokeShell(high, weigh)
}

func (i iShell) ResizePty(high, weigh int) error {
	req := ptyWindowChangeMsg{
		Columns: uint32(weigh),
		Rows:    uint32(high),
		Width:   uint32(weigh * 8),
		Height:  uint32(high * 8),
	}
	_, err := i.ch.SendRequest("window-change", false, ssh.Marshal(&req))
	return err
}

func (i iShell) ChanSend(cmd string) error {
	_, err := i.ch.Write([]byte(cmd))
	if err == io.EOF {
		i.ChanClose()
	}
	return err
}

func (i iShell) ChanRcv(ch chan string) {
	br := bufio.NewReader(i.ch)
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			close(ch)
			break
		}
		ch <- string(r)
		time.Sleep(100 * time.Microsecond)
	}
}

func (i *iShell) ChanClose() {
	i.ch.Close()
}
