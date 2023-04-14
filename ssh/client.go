package ssh

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slog"
)

type Client struct {
	// User Authenticate as.
	User string `yaml:"user"`

	// Host SSH server to connect to.
	Host string `yaml:"host"`

	// Password Used for SSH authentication.
	Password string `yaml:"password,omitempty"`

	// PrivateKey Used for SSH authentication.
	PrivateKey string `yaml:"private_key,omitempty"`

	// Passphrase Private key passphrase.
	Passphrase string `yaml:"passphrase,omitempty"`

	// LogLevel Level of logging print.
	LogLevel int8 `yaml:"log_level,omitempty"`

	// LogEncoding Log output format.
	LogEncoding string `yaml:"log_encoding"`

	// RetryMin Minimum time to retry connecting to the ssh server
	RetryMin time.Duration `yaml:"retry_min,omitempty"`

	// RetryMax Maximum time to retry connecting to the ssh server
	RetryMax time.Duration `yaml:"retry_max,omitempty"`

	// ServerAliveInterval Interval to use for the ssh server's keepalive
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`

	// ServerAliveCountMax Maximum number of keepalive packets to send
	ServerAliveCountMax uint32 `yaml:"server_alive_count_max"`

	// Logger Used for logging
	Logger *slog.Logger
}

func (c *Client) String() string {
	return fmt.Sprintf("%s@%s", c.User, c.Host)
}

func (c *Client) Validate() error {
	if c.PrivateKey == "" && c.Password == "" {
		return fmt.Errorf("one of [password, private_key] required")
	}

	if c.RetryMin <= 0 {
		c.RetryMin = time.Second
	}

	if c.RetryMax <= 0 {
		c.RetryMax = time.Minute
	}

	if c.ServerAliveInterval <= 0 {
		c.ServerAliveInterval = 0
	}

	if c.ServerAliveCountMax <= 1 {
		c.ServerAliveCountMax = 3
	}

	return nil
}

func (c *Client) keepAlive(ctx context.Context, conn ssh.Conn, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			_, _, err := conn.SendRequest("keepalive@batproxy.dev", true, nil)
			if err != nil {
				c.Logger.Error("keepalive", "err", err)
				return
			}
			c.Logger.Debug("keepalive",
				"client", c,
			)
		case <-ctx.Done():
			c.Logger.Debug("keepalive",
				"status", "exit",
				"client", c,
			)
			return
		}
	}
}

func (c *Client) getCurrentTempDelay(tempDelay time.Duration) time.Duration {
	if tempDelay == 0 {
		tempDelay = c.RetryMin
	} else {
		tempDelay *= 2
	}
	if tempDelay > c.RetryMax {
		tempDelay = c.RetryMax
	}

	return tempDelay
}
