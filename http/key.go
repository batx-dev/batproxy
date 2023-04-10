package http

import (
	"context"
	"fmt"

	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/memo"
	"github.com/batx-dev/batproxy/ssh"
)

type key struct {
	host         string
	user         string
	identityFile string
	password     string
}

func (k *key) String() string {
	return fmt.Sprintf("%s@%s", k.user, k.host)
}

func dialFunc(logger logger.Logger) memo.Func[key, *ssh.Ssh] {
	return func(ctx context.Context, key key, cleanup func()) (*ssh.Ssh, error) {
		client := &ssh.Client{
			Host:         key.host,
			User:         key.user,
			IdentityFile: key.identityFile,
			Password:     key.password,
		}

		s := ssh.New(logger, client)
		return s, nil
	}
}
