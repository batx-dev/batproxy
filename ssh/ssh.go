package ssh

import (
	"context"
	"net"

	"github.com/batx-dev/batproxy/logger"
)

type Ssh struct {
	conn net.Conn

	Client *Client `yaml:"client"`

	Logger logger.Logger

	forwards map[string]string
}

func (s *Ssh) New() (*Ssh, error) {
	client := &Client{}
	l := s.Logger

	if client.LogLevel != 0 {
		l.LogLevel = client.LogLevel
	}
	if client.LogEncoding != "" {
		l.Encoding = client.LogEncoding
	}

	client.Logger = l.Build().WithName("ssh")

	return &Ssh{
		Client: client,

		forwards: make(map[string]string, 10),
	}, nil
}

func (s *Ssh) Run(ctx context.Context) error {
	go func(c *Client) {
		l := s.Logger

		if c.LogLevel != 0 {
			l.LogLevel = c.LogLevel
		}
		if c.LogEncoding != "" {
			l.Encoding = c.LogEncoding
		}

		c.Logger = l.Build().WithName("ssh")

		if err := c.Dial(ctx); err != nil {
			c.Logger.Error(err, "dial", "host", c.Host)
			return
		}
	}(s.Client)
	return nil
}

func (s *Ssh) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	c := s.Client

	if s.conn != nil {
		return s.conn, nil
	}

	dialDestination, err := c.dialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	return dialDestination, nil
}
