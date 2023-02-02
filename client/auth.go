package client

import (
	"os"

	"golang.org/x/crypto/ssh"
)

type AuthMethod func() (ssh.AuthMethod, error)

func PasswordAuth(password string) AuthMethod {
	return func() (ssh.AuthMethod, error) {
		return ssh.Password(password), nil
	}
}

func PrivateKeyAuth(pemBytes []byte) AuthMethod {
	return func() (ssh.AuthMethod, error) {
		if signer, err := ssh.ParsePrivateKey(pemBytes); err != nil {
			return nil, err
		} else {
			return ssh.PublicKeys(signer), nil
		}
	}
}

func PrivateKeyWithPassphraseAuth(pemBytes, passphrase []byte) AuthMethod {
	return func() (ssh.AuthMethod, error) {
		if signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, passphrase); err != nil {
			return nil, err
		} else {
			return ssh.PublicKeys(signer), nil
		}
	}
}

func PrivateKeyFileAuth(keyFile string) AuthMethod {
	return func() (ssh.AuthMethod, error) {
		if f, err := os.ReadFile(keyFile); err != nil {
			return nil, err
		} else {
			return PrivateKeyAuth(f)()
		}
	}
}

func PrivateKeyFileWithPassphraseAuth(keyFile string, passphrase []byte) AuthMethod {
	return func() (ssh.AuthMethod, error) {
		if f, err := os.ReadFile(keyFile); err != nil {
			return nil, err
		} else {
			ssh.PublicKeys()
			return PrivateKeyWithPassphraseAuth(f, passphrase)()
		}
	}
}
