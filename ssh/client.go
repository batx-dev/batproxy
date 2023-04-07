package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/crypto/ssh"
	"inet.af/tcpproxy"
)

type Client struct {
	// ln is the local listener
	ln net.Listener

	Conn *ssh.Client

	// Host is the ssh server to connect to.
	Host string `yaml:"host"`

	// User is the user to authenticate as.
	User string `yaml:"user"`

	// Password is the password to use for authentication.
	Password string `yaml:"password,omitempty"`

	// IdentityFile is the path to the private key file to use for authentication.
	IdentityFile string `yaml:"identity_file,omitempty"`

	// LogLevel is the log level to use for logging
	LogLevel int8 `yaml:"log_level,omitempty"`

	// LogEncoding is the log output format
	LogEncoding string `yaml:"log_encoding"`

	// RetryMin is the minimum time to retry connecting to the ssh server
	RetryMin time.Duration `yaml:"retry_min,omitempty"`

	// RetryMax is the maximum time to retry connecting to the ssh server
	RetryMax time.Duration `yaml:"retry_max,omitempty"`

	// ServerAliveInterval is the interval to use for the ssh server's keepalive
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`

	// ServerAliveCountMax is the maximum number of keepalive packets to send
	ServerAliveCountMax uint32 `yaml:"server_alive_count_max"`

	// Rules is the list of rules to use for the ssh server
	Rules []Rule `yaml:"rules"`

	// Logger is the logger to use for logging
	Logger logr.Logger
}

type Rule struct {
	// Remote is the remote address to forward to
	Remote string `yaml:"remote,omitempty"`

	// Local is the local address to forward to
	Local string `yaml:"local,omitempty"`

	// Reverse is whether to reverse the direction of the connection
	Reverse bool `yaml:"reverse,omitempty"`
}

func (c *Client) Dial(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}

	var auth []ssh.AuthMethod
	if c.IdentityFile != "" {
		key, err := os.ReadFile(c.IdentityFile)
		if err != nil {
			return err
		}
		singer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return err
		}
		auth = append(auth, ssh.PublicKeys(singer))
	}
	if c.Password != "" {
		auth = append(auth, ssh.Password(c.Password))
	}

	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 30,
	}

	// how long to sleep on accept failure
	var tempDelay time.Duration
	var tempConnDelay time.Duration
	for {
		// establish connect with remote host
		conn, err := DialTimeout("tcp",
			c.Host,
			config,
			c.ServerAliveInterval*time.Duration(c.ServerAliveCountMax),
		)
		if err != nil {
			tempDelay = c.getCurrentTempDelay(tempDelay)
			c.Logger.Error(err, "dial",
				"status", "retry",
				"host", c.Host,
				"user", c.User,
				"retry_in", tempDelay,
			)

			select {
			case <-ctx.Done():
				c.Logger.Error(err, "dial",
					"status", "cancel",
					"host", c.Host,
					"user", c.User,
				)
				return nil
			case <-time.After(tempDelay):
				continue
			}
		}
		tempDelay = 0
		startTime := time.Now()

		c.Logger.V(1).Info("dial",
			"status", "ok",
			"host", c.Host,
			"user", c.User,
		)

		connErr := make(chan error, 1)
		go func() {
			connErr <- conn.Wait()
		}()

		// Listen remote listen and forward for each rule
		childCtx, childCancel := context.WithCancel(ctx)
		if c.ServerAliveInterval > 0 {
			go c.keepAlive(childCtx, conn, c.ServerAliveInterval)
		}

		go c.Listen(childCtx, conn)

		select {
		case <-ctx.Done():
			childCancel()
			conn.Close()
			return nil
		case err := <-connErr:
			c.Logger.Error(err, "connect",
				"status", "retry",
				"host", c.Host,
				"user", c.User,
				"retry_in", tempConnDelay,
			)
			childCancel()
			conn.Close()

			// avoid too frequent reconnection
			durationTime := time.Since(startTime)
			tempConnDelay = c.getCurrentTempDelay(tempConnDelay)
			if durationTime < tempConnDelay {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(tempConnDelay - durationTime):
					continue
				}
			}
			if durationTime > c.RetryMax {
				tempConnDelay = 0
			}
		}
	}
}

func (c *Client) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	var auth []ssh.AuthMethod
	if c.IdentityFile != "" {
		key, err := os.ReadFile(c.IdentityFile)
		if err != nil {
			return nil, err
		}
		singer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(singer))
	}
	if c.Password != "" {
		auth = append(auth, ssh.Password(c.Password))
	}

	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 30,
	}

	// how long to sleep on accept failure
	var tempDelay time.Duration
	for {
		// establish connect with remote host
		conn, err := DialTimeout("tcp",
			c.Host,
			config,
			c.ServerAliveInterval*time.Duration(c.ServerAliveCountMax),
		)
		if err != nil {
			tempDelay = c.getCurrentTempDelay(tempDelay)
			c.Logger.Error(err, "dial",
				"status", "retry",
				"host", c.Host,
				"user", c.User,
				"retry_in", tempDelay,
			)

			select {
			case <-ctx.Done():
				c.Logger.Error(err, "dial",
					"status", "cancel",
					"host", c.Host,
					"user", c.User,
				)
				return nil, nil
			case <-time.After(tempDelay):
				continue
			}
		}
		tempDelay = 0

		return conn.Dial(network, address)
	}

}

// Listen address which define in rules
func (c *Client) Listen(ctx context.Context, conn *ssh.Client) {
	for _, rule := range c.Rules {
		if rule.Reverse {
			go c.listenRemote(ctx, conn, rule.Local, rule.Remote, rule.Reverse)
		} else {
			go c.listenLocal(ctx, conn, rule.Local, rule.Remote, rule.Reverse)
		}
	}
}

func (c *Client) listenRemote(ctx context.Context, conn *ssh.Client, localAddr, remoteAddr string, reverse bool) {
	var tempDelay time.Duration // how long to sleep on accept failure
ListenRemote:
	for {
		listener, err := conn.Listen("tcp", remoteAddr)
		if err != nil {
			tempDelay = c.getCurrentTempDelay(tempDelay)

			c.Logger.Error(err, "listen",
				"status", "retry",
				"host", c.Host,
				"user", c.User,
				"remote", remoteAddr,
				"reverse", reverse,
				"retry_in", tempDelay,
			)

			select {
			case <-ctx.Done():
				c.Logger.V(2).Info("listen",
					"status", "cancel",
					"host", c.Host,
					"user", c.User,
					"remote", remoteAddr,
					"reverse", reverse,
				)
				return
			case <-time.After(tempDelay):
				continue
			}
		}
		tempDelay = 0

		c.Logger.V(1).Info("listen",
			"status", "ok",
			"host", c.Host,
			"user", c.User,
			"remote", remoteAddr,
			"reverse", reverse,
		)

		// accept message and forward
		go c.forward(ctx, listener, localAddr, reverse, nil)

		select {
		case <-ctx.Done():
			c.Logger.V(1).Info("listen",
				"status", "cancel",
				"host", c.Host,
				"user", c.User,
				"remote", remoteAddr,
				"reverse", reverse,
			)
			listener.Close()
			break ListenRemote
		}
	}
}

func (c *Client) listenLocal(ctx context.Context, conn *ssh.Client, localAddr, remoteAddr string, reverse bool) {
	var tempDelay time.Duration // how long to sleep on accept failure
ListenLocal:
	for {
		listener, err := net.Listen("tcp", localAddr)
		if err != nil {
			tempDelay = c.getCurrentTempDelay(tempDelay)

			c.Logger.Error(err, "listen",
				"status", "retry",
				"host", c.Host,
				"user", c.User,
				"local", localAddr,
				"reverse", reverse,
				"retry_in", tempDelay,
			)

			select {
			case <-ctx.Done():
				c.Logger.V(2).Info("listen",
					"status", "cancel",
					"host", c.Host,
					"user", c.User,
					"local", localAddr,
					"reverse", reverse,
				)
				return
			case <-time.After(tempDelay):
				continue
			}
		}
		tempDelay = 0

		c.Logger.V(1).Info("listen",
			"status", "ok",
			"host", c.Host,
			"user", c.User,
			"local", localAddr,
			"reverse", reverse,
		)

		// accept message and forward
		go c.forward(ctx, listener, remoteAddr, false, func(ctx context.Context, network, address string) (net.Conn, error) {
			return conn.Dial(network, address)
		})

		select {
		case <-ctx.Done():
			c.Logger.V(1).Info("listen",
				"status", "cancel",
				"host", c.Host,
				"user", c.User,
				"local", localAddr,
				"reverse", reverse,
			)
			listener.Close()
			break ListenLocal
		}
	}
}

func (c *Client) forward(ctx context.Context, source net.Listener, destination string, reverse bool, dialFunc func(ctx context.Context, network, address string) (net.Conn, error)) {
	// default local's value is rule remote address when reverse is true
	local := destination
	remote := source.Addr().String()
	dialProxy := tcpproxy.To(destination)
	dialProxy.DialTimeout = time.Second * 15
	if !reverse {
		dialProxy.DialContext = dialFunc
		local, remote = remote, local
	}

	dialProxy.OnDialError = func(src net.Conn, dstDialErr error) {
		c.Logger.Error(dstDialErr, "forward",
			"status", "failure",
			"host", c.Host,
			"user", c.User,
			"remote", remote,
			"local", local,
			"reverse", reverse,
		)
	}

	var tempDelay time.Duration
	for {
		accept, err := source.Accept()
		if err != nil {
			tempDelay = c.getCurrentTempDelay(tempDelay)

			select {
			case <-ctx.Done():
				c.Logger.Error(err, "accept",
					"status", "cancel",
					"host", c.Host,
					"user", c.User,
					"remote", remote,
					"local", local,
					"reverse", reverse,
				)
				return
			case <-time.After(tempDelay):
				c.Logger.Error(err, "accept",
					"status", "retry",
					"host", c.Host,
					"user", c.User,
					"remote", remote,
					"local", local,
					"reverse", reverse,
					"retry_in", tempDelay,
				)
				continue
			}
		}
		tempDelay = 0

		c.Logger.V(2).Info("forward",
			"status", "start",
			"host", c.Host,
			"user", c.User,
			"remote", remote,
			"local", local,
			"reverse", reverse,
		)
		go func() {
			dialProxy.HandleConn(accept)
			c.Logger.V(2).Info("forward",
				"status", "end",
				"host", c.Host,
				"user", c.User,
				"remote", remote,
				"local", local,
				"reverse", reverse,
			)
		}()
	}
}

func (c *Client) Validate() error {
	if c.IdentityFile == "" && c.Password == "" {
		return fmt.Errorf("one of [password, identity_file] required")
	}

	if c.IdentityFile != "" {
		if strings.HasPrefix(c.IdentityFile, "~") {
			homePath, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			c.IdentityFile = strings.Replace(c.IdentityFile, "~", homePath, 1)
		}
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
			_, _, err := conn.SendRequest("keepalive@psh.dev", true, nil)
			if err != nil {
				c.Logger.Error(err, "keepalive")
				return
			}
			c.Logger.V(2).Info("keepalive",
				"host", c.Host,
				"user", c.User,
			)
		case <-ctx.Done():
			c.Logger.V(2).Info("keepalive",
				"status", "exit",
				"host", c.Host,
				"user", c.User,
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
