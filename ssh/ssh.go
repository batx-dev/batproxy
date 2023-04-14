package ssh

import (
	"context"
	"net"

	"github.com/batx-dev/batproxy/memo"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slog"
)

type Ssh struct {
	memo *memo.Memo[key, *ssh.Client]

	Client *Client `yaml:"client"`

	Logger *slog.Logger
}

func New(logger *slog.Logger, client *Client) *Ssh {
	return &Ssh{
		Client: client,
		Logger: logger,
		memo:   memo.New(dialFunc(client)),
	}
}

func (s *Ssh) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	sc, err := s.memo.Get(ctx, key{
		User: s.Client.User,
		Host: s.Client.Host,
	})
	if err != nil {
		return nil, err
	}
	return sc.Dial(network, address)
}
