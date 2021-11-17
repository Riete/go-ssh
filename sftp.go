package go_ssh

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
)

func (h *Host) Put(files []FilePut) error {
	if err := h.openSftp(); err != nil {
		return err
	}
	defer h.SSHClient.Close()
	for _, v := range files {
		if err := h.put(v.LocalFile, v.RemoteDir); err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) Get(files []FileGet) error {
	if err := h.openSftp(); err != nil {
		return err
	}
	defer h.SSHClient.Close()
	for _, v := range files {
		if err := h.get(v.LocalDir, v.RemoteFile); err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) openSftp() error {
	if err := h.connect(); err != nil {
		return err
	}
	if sftpClient, err := sftp.NewClient(h.SSHClient); err != nil {
		return errors.New(fmt.Sprintf("open sftp failed, %s", err.Error()))
	} else {
		h.SFTPClient = sftpClient
	}
	return nil
}

func (h *Host) put(local, remote string) error {
	localFile, err := os.Open(local)
	if err != nil {
		return errors.New(fmt.Sprintf("Open local file %s failed: %s", local, err.Error()))
	}
	filename := path.Base(local)
	remotePath := path.Join(remote, filename)
	remoteFile, err := h.SFTPClient.Create(remotePath)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s]: Create remote file %s failed: %s", h.IpAddr, remotePath, err.Error()))
	}
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s]: Upload file to %s failed: %s", h.IpAddr, remotePath, err.Error()))
	}
	return nil
}

func (h *Host) get(local, remote string) error {
	filename := path.Base(remote)
	localPath := path.Join(local, filename)
	localFile, err := os.Create(localPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Create local file %s failed: %s", localPath, err.Error()))
	}
	remoteFile, err := h.SFTPClient.Open(remote)
	if err != nil {
		return errors.New(fmt.Sprintf("Open remote file %s failed: %s", remote, err.Error()))
	}
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s]: Download file to %s failed: %s", h.IpAddr, localPath, err.Error()))
	}
	return nil
}
