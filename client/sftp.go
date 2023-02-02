package client

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
)

type sftpServer struct {
	*sshServer
	sftpClient *sftp.Client
}

type SFTPExecutor interface {
	Get(remotePath string) (*sftp.File, error)
	GetAndSave(remotePath string, writer io.WriteCloser) error
	Put(reader io.ReadCloser, remotePath string) error
	LsDir(remoteDir string) ([]os.FileInfo, error)
	RawClient() *sftp.Client
	Close()
	Open() error
}

func NewSFTPExecutor(username, password, ipaddr, port, privateKeyPath string) SFTPExecutor {
	return &sftpServer{sshServer: newSSHServer(username, password, ipaddr, port, privateKeyPath)}
}

func (sf *sftpServer) Open() error {
	return sf.openSftp()
}

func (sf sftpServer) LsDir(remoteDir string) ([]os.FileInfo, error) {
	return sf.sftpClient.ReadDir(remoteDir)
}

func (sf sftpServer) GetAndSave(remotePath string, writer io.WriteCloser) error {
	defer writer.Close()
	reader, err := sf.Get(remotePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if _, err := io.Copy(writer, reader); err != nil {
		return errors.New(fmt.Sprintf("save local file failed: %s", err.Error()))
	}
	return nil
}

func (sf *sftpServer) Get(remoteFile string) (*sftp.File, error) {
	return sf.get(remoteFile)
}

func (sf *sftpServer) Put(reader io.ReadCloser, remotePath string) error {
	return sf.put(reader, remotePath)
}

func (sf *sftpServer) openSftp() error {
	var err error
	if err = sf.connect(); err != nil {
		return err
	}
	if sf.sftpClient, err = sftp.NewClient(sf.sshClient); err != nil {
		return errors.New(fmt.Sprintf("open sftp failed: %s", err.Error()))
	}
	return nil
}

func (sf *sftpServer) put(reader io.ReadCloser, remotePath string) error {
	defer reader.Close()
	writer, err := sf.sftpClient.Create(remotePath)
	if err != nil {
		return errors.New(fmt.Sprintf("create remote file failed: %s", err.Error()))
	}
	defer writer.Close()
	if _, err = io.Copy(writer, reader); err != nil {
		return errors.New(fmt.Sprintf("save remote file failed: %s", err.Error()))
	}
	return nil
}

func (sf *sftpServer) get(remoteFile string) (*sftp.File, error) {
	file, err := sf.sftpClient.Open(remoteFile)
	if err != nil {
		return file, errors.New(fmt.Sprintf("open remote file failed: %s", err.Error()))
	}
	return file, nil
}

func (sf *sftpServer) RawClient() *sftp.Client {
	return sf.sftpClient
}

func (sf *sftpServer) Close() {
	_ = sf.sftpClient.Close()
	_ = sf.sshClient.Close()
}
