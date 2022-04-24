package go_ssh

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
)

type FileGet struct {
	LocalDir   string
	RemoteFile string
}

type FilePut struct {
	LocalFile string
	RemoteDir string
	Src       io.ReadCloser
}

type sftpServer struct {
	*sshServer
	sftpClient *sftp.Client
}

type SFTPExecutor interface {
	Get(file FileGet) error
	Put(file FilePut) error
	BatchGet(files []FileGet) error
	BatchPut(files []FilePut) error
	RawClient() *sftp.Client
}

func NewSFTPExecutor(username, password, ipaddr, port string) SFTPExecutor {
	return &sftpServer{sshServer: newSSHServer(username, password, ipaddr, port)}
}

func (sf *sftpServer) Get(file FileGet) error {
	return sf.BatchGet([]FileGet{file})
}

func (sf *sftpServer) Put(file FilePut) error {
	return sf.BatchPut([]FilePut{file})
}

func (sf *sftpServer) BatchPut(files []FilePut) error {
	if err := sf.openSftp(); err != nil {
		return err
	}
	defer sf.sshClient.Close()
	for _, v := range files {
		if err := sf.put(v.LocalFile, v.Src, v.RemoteDir); err != nil {
			return err
		}
	}
	return nil
}

func (sf *sftpServer) BatchGet(files []FileGet) error {
	if err := sf.openSftp(); err != nil {
		return err
	}
	defer sf.sshClient.Close()
	for _, v := range files {
		if err := sf.get(v.LocalDir, v.RemoteFile); err != nil {
			return err
		}
	}
	return nil
}

func (sf *sftpServer) openSftp() error {
	if err := sf.connect(); err != nil {
		return err
	}
	if sftpClient, err := sftp.NewClient(sf.sshClient); err != nil {
		return errors.New(fmt.Sprintf("open sftp failed, %s", err.Error()))
	} else {
		sf.sftpClient = sftpClient
	}
	return nil
}

func (sf *sftpServer) put(local string, src io.ReadCloser, remote string) error {
	if src == nil {
		var err error
		if src, err = os.Open(local); err != nil {
			return errors.New(fmt.Sprintf("Open local file %s failed: %s", local, err.Error()))
		}
	}
	defer src.Close()
	filename := path.Base(local)
	remotePath := path.Join(remote, filename)
	remoteFile, err := sf.sftpClient.Create(remotePath)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s]: Create remote file %s failed: %s", sf.ipaddr, remotePath, err.Error()))
	}
	defer remoteFile.Close()
	_, err = io.Copy(remoteFile, src)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s]: Upload file to %s failed: %s", sf.ipaddr, remotePath, err.Error()))
	}
	return nil
}

func (sf *sftpServer) get(local, remote string) error {
	filename := path.Base(remote)
	localPath := path.Join(local, filename)
	remoteFile, err := sf.sftpClient.Open(remote)
	if err != nil {
		return errors.New(fmt.Sprintf("Open remote file %s failed: %s", remote, err.Error()))
	}
	defer remoteFile.Close()
	localFile, err := os.Create(localPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Create local file %s failed: %s", localPath, err.Error()))
	}
	defer localFile.Close()
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s]: Download file to %s failed: %s", sf.ipaddr, localPath, err.Error()))
	}
	return nil
}

func (sf *sftpServer) RawClient() *sftp.Client {
	return sf.sftpClient
}
