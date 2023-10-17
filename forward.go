package go_ssh

import (
	"context"
	"io"
	"net"

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

// dailRemote let remote peer dail host:port
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
	defer local.Close()
	defer remote.Close()
	var err error
	go func() {
		if _, err1 := io.Copy(local, remote); err1 != nil {
			err = err1
		}
	}()
	_, err2 := io.Copy(remote, local)
	if err != nil {
		return err
	}
	return err2
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
	ch := make(chan error)
	localListener, err := l.ListenLocal()
	if err != nil {
		return err
	}
	defer localListener.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-ch:
			return err
		default:
			local, err := localListener.Accept()
			if err != nil {
				return err
			}
			go func() {
				remote, err := l.DailRemote()
				if err != nil {
					ch <- err
				}
				if err := l.f.forward(local, remote); err != nil {
					ch <- err
				}
			}()
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
	ch := make(chan error)
	remoteListener, err := r.ListenRemote()
	if err != nil {
		return err
	}
	defer remoteListener.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-ch:
			return err
		default:
			remote, err := remoteListener.Accept()
			if err != nil {
				return err
			}
			go func() {
				local, err := r.DailLocal()
				if err != nil {
					ch <- err
				}
				if err := r.f.forward(local, remote); err != nil {
					ch <- err
				}
			}()
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
