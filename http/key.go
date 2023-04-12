package http

import (
	"context"
	"fmt"

	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/memo"
	"github.com/batx-dev/batproxy/ssh"
)

type key struct {
	// User Over SSH login name.
	// Required.
	User string `json:"user"`

	// Host Over SSH login host.
	// Required.
	Host string `json:"host"`

	// PrivateKey Over SSH login private key.
	// Optional.
	PrivateKey string `json:"private_key,omitempty"`

	// Passphrase Over SSH login private key password.
	// Optional.
	Passphrase string `json:"passphrase,omitempty"`

	// Password Over SSH login password.
	// Optional.
	Password string `json:"password,omitempty"`
}

func (k *key) String() string {
	return fmt.Sprintf("%s@%s", k.User, k.Host)
}

func sshFunc(logger logger.Logger) memo.Func[key, *ssh.Ssh] {
	return func(ctx context.Context, key key, cleanup func()) (*ssh.Ssh, error) {
		client := &ssh.Client{
			User:       key.User,
			Host:       key.Host,
			PrivateKey: key.PrivateKey,
			Passphrase: key.Passphrase,
			Password:   key.Password,
		}

		s := ssh.New(logger, client)
		return s, nil
	}
}
