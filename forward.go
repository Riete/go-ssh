package go_ssh

import (
	"context"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type PortForwarder interface {
	PortForward(context.Context) error
}

type portForward struct {
	client *ssh.Client
}

// listenLocal listen at host:port
func (p portForward) listenLocal(host, port string) (net.Listener, error) {
	return net.Listen("tcp", host+":"+port)
}

// dailRemote dail to remote peer's host:port
func (p portForward) dailRemote(host, port string) (net.Conn, error) {
	return p.client.Dial("tcp", host+":"+port)
}

// listenRemote let remote peer listen at host:port
func (p portForward) listenRemote(host, port string) (net.Listener, error) {
	return p.client.Listen("tcp", host+":"+port)
}

// dailLocal dail host:port
func (p portForward) dailLocal(host, port string) (net.Conn, error) {
	return net.Dial("tcp", host+":"+port)
}

func (p portForward) forward(local, remote net.Conn) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	var err error
	go func() {
		defer wg.Done()
		_, err = io.Copy(local, remote)
	}()
	go func() {
		defer wg.Done()
		_, err = io.Copy(remote, local)
	}()
	return err
}

type LocalToRemote struct {
	f          *portForward
	localHost  string
	localPort  string
	remoteHost string
	remotePort string
}

func (l LocalToRemote) ListenLocal() (net.Listener, error) {
	return l.f.listenLocal(l.localHost, l.localPort)
}

func (l LocalToRemote) DailRemote() (net.Conn, error) {
	return l.f.dailRemote(l.remoteHost, l.remotePort)
}

// PortForward  local host:port -> remote host:port
func (l LocalToRemote) PortForward(ctx context.Context) error {
	var localListener net.Listener
	var remote net.Conn
	var local net.Conn
	var err error

	defer func() {
		if remote != nil {
			_ = remote.Close()
		}
		if local != nil {
			_ = local.Close()
		}
		if localListener != nil {
			_ = localListener.Close()
		}
	}()

	localListener, err = l.ListenLocal()
	if err != nil {
		return err
	}
	remote, err = l.DailRemote()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			local, err = localListener.Accept()
			if err != nil {
				return err
			}
			if err = l.f.forward(local, remote); err != nil {
				return err
			}
		}
	}
}

func NewLocalToRemoteForward(client *ssh.Client, localHost, localPort, remoteHost, remotePort string) PortForwarder {
	return &LocalToRemote{
		f:          &portForward{client: client},
		localHost:  localHost,
		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,
	}
}

type RemoteToLocal struct {
	f          *portForward
	localHost  string
	localPort  string
	remoteHost string
	remotePort string
}

func (r RemoteToLocal) ListenRemote() (net.Listener, error) {
	return r.f.listenRemote(r.remoteHost, r.remotePort)
}

func (r RemoteToLocal) DailLocal() (net.Conn, error) {
	return r.f.dailLocal(r.localHost, r.localPort)
}

// PortForward  remote host:port -> local host:port
func (r RemoteToLocal) PortForward(ctx context.Context) error {
	var remoteListener net.Listener
	var remote net.Conn
	var local net.Conn
	var err error

	defer func() {
		if remote != nil {
			_ = remote.Close()
		}
		if local != nil {
			_ = local.Close()
		}
		if remoteListener != nil {
			_ = remoteListener.Close()
		}
	}()

	remoteListener, err = r.ListenRemote()
	if err != nil {
		return err
	}
	local, err = r.DailLocal()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			remote, err = remoteListener.Accept()
			if err != nil {
				return err
			}
			if err = r.f.forward(local, remote); err != nil {
				return err
			}
		}
	}
}

func NewRemoteToLocalForward(client *ssh.Client, localHost, localPort, remoteHost, remotePort string) PortForwarder {
	return &RemoteToLocal{
		f:          &portForward{client: client},
		localHost:  localHost,
		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,
	}
}
