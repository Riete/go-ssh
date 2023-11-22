package ssh

import (
	"context"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type portForward struct {
	client *ssh.Client
}

// listenLocal listen at addr
func (p portForward) listenLocal(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

// dailRemote let remote peer dail addr
func (p portForward) dailRemote(addr string) (net.Conn, error) {
	return p.client.Dial("tcp", addr)
}

// listenRemote let remote peer listen at addr
func (p portForward) listenRemote(addr string) (net.Listener, error) {
	return p.client.Listen("tcp", addr)
}

// dailLocal dail addr
func (p portForward) dailLocal(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

func (p portForward) forward(src, dst io.ReadWriteCloser) {
	var wg sync.WaitGroup
	var o sync.Once
	closeReader := func() {
		_ = src.Close()
		_ = dst.Close()
	}

	wg.Add(2)
	go func() {
		io.Copy(src, dst)
		o.Do(closeReader)
		wg.Done()
	}()

	go func() {
		io.Copy(dst, src)
		o.Do(closeReader)
		wg.Done()
	}()
	wg.Wait()
}

// PortForwarder
// local and remote is ip:port
type PortForwarder struct {
	f      *portForward
	local  string
	remote string
}

func (p PortForwarder) LocalForwardToRemote() error {
	return p.LocalForwardToRemoteContext(context.Background())
}

func (p PortForwarder) LocalForwardToRemoteContext(ctx context.Context) error {
	l, err := p.f.listenLocal(p.local)
	if err != nil {
		return err
	}
	defer l.Close()
	ch := make(chan error)

	go func() {
		for {
			local, err := l.Accept()
			if err != nil {
				ch <- err
				return
			}
			remote, err := p.f.dailRemote(p.remote)
			if err != nil {
				ch <- err
				return
			}
			go p.f.forward(local, remote)
		}
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-ch:
		return err
	}
}

func (p PortForwarder) RemoteForwardToLocal() error {
	return p.RemoteForwardToLocalContext(context.Background())
}

func (p PortForwarder) RemoteForwardToLocalContext(ctx context.Context) error {
	r, err := p.f.listenRemote(p.remote)
	if err != nil {
		return err
	}
	defer r.Close()
	ch := make(chan error)

	go func() {
		for {
			remote, err := r.Accept()
			if err != nil {
				ch <- err
				return
			}
			local, err := p.f.dailLocal(p.local)
			if err != nil {
				ch <- err
				return
			}
			go p.f.forward(remote, local)
		}
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-ch:
		return err
	}

}

func NewPortForwarder(client *ssh.Client, local, remote string) *PortForwarder {
	return &PortForwarder{f: &portForward{client: client}, local: local, remote: remote}
}
