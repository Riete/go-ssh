package go_ssh

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Host struct {
	Username   string
	Password   string
	IpAddr     string
	Port       string
	SSHClient  *ssh.Client
	SFTPClient *sftp.Client
}

func NewHost(username, password, ipaddr, port string) Host {
	return Host{Username: username, Password: password, IpAddr: ipaddr, Port: port}
}

type FileGet struct {
	LocalDir   string
	RemoteFile string
}

type FilePut struct {
	LocalFile string
	RemoteDir string
}
