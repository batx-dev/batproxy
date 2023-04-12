package ssh

import (
	"context"
	"net"

	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/memo"
	"golang.org/x/crypto/ssh"
)

type Ssh struct {
	memo *memo.Memo[key, *ssh.Client]

	Client *Client `yaml:"client"`

	Logger logger.Logger
}

func New(logger logger.Logger, client *Client) *Ssh {
	if client.LogLevel != 0 {
		logger.LogLevel = client.LogLevel
	}
	if client.LogEncoding != "" {
		logger.Encoding = client.LogEncoding
	}

	client.Logger = logger.Build().WithName("ssh")

	return &Ssh{
		Client: client,
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
