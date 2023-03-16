package go_ssh

import (
	"errors"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/pkg/sftp"
)

type SftpClient struct {
	server     *Server
	client     *ssh.Client
	sftpClient *sftp.Client
}

type SftpExecutor interface {
	Connect(username string, timeout time.Duration, methods ...AuthMethod) error
	Get(remotePath string) (*sftp.File, error)
	GetAndSave(remotePath string, writer io.WriteCloser) error
	Put(reader io.ReadCloser, remotePath string) error
	LsDir(remoteDir string) ([]os.FileInfo, error)
	RawClient() *sftp.Client
	Close()
}

func (s *SftpClient) Connect(username string, timeout time.Duration, methods ...AuthMethod) error {
	var err error
	if s.client, err = s.server.connect(username, timeout, methods...); err != nil {
		return errors.New("connect to server failed: " + err.Error())
	}
	return s.openSftp()
}

func (s SftpClient) LsDir(remoteDir string) ([]os.FileInfo, error) {
	return s.sftpClient.ReadDir(remoteDir)
}

func (s SftpClient) GetAndSave(remotePath string, writer io.WriteCloser) error {
	defer writer.Close()
	reader, err := s.Get(remotePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if _, err := io.Copy(writer, reader); err != nil {
		return errors.New("save local file failed: " + err.Error())
	}
	return nil
}

func (s *SftpClient) Get(remoteFile string) (*sftp.File, error) {
	return s.get(remoteFile)
}

func (s *SftpClient) Put(reader io.ReadCloser, remotePath string) error {
	return s.put(reader, remotePath)
}

func (s *SftpClient) openSftp() error {
	var err error
	if s.sftpClient, err = sftp.NewClient(s.client); err != nil {
		return errors.New("open sftp failed: " + err.Error())
	}
	return nil
}

func (s *SftpClient) put(reader io.ReadCloser, remotePath string) error {
	defer reader.Close()
	writer, err := s.sftpClient.Create(remotePath)
	if err != nil {
		return errors.New("create remote file failed: " + err.Error())
	}
	defer writer.Close()
	if _, err = io.Copy(writer, reader); err != nil {
		return errors.New("save remote file failed: " + err.Error())
	}
	return nil
}

func (s *SftpClient) get(remoteFile string) (*sftp.File, error) {
	file, err := s.sftpClient.Open(remoteFile)
	if err != nil {
		return file, errors.New("open remote file failed: " + err.Error())
	}
	return file, nil
}

func (s *SftpClient) RawClient() *sftp.Client {
	return s.sftpClient
}

func (s *SftpClient) Close() {
	_ = s.sftpClient.Close()
	_ = s.client.Close()
}

func NewSftpExecutor(server *Server) SftpExecutor {
	return &SftpClient{server: server}
}
