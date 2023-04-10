package ssh

import (
	"context"
	"os"
	"time"

	"github.com/batx-dev/batproxy/memo"
	"golang.org/x/crypto/ssh"
)

func dialFunc() memo.Func[*Client, *ssh.Client] {
	return func(ctx context.Context, c *Client, cleanup func()) (*ssh.Client, error) {
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

		// establish connect with remote host
		client, err := DialTimeout("tcp",
			c.Host,
			config,
			c.ServerAliveInterval*time.Duration(c.ServerAliveCountMax),
		)
		if err != nil {
			c.Logger.Error(err, "dial",
				"status", "fail",
				"host", c.Host,
				"user", c.User,
			)
			go func() {
				time.Sleep(15 * time.Second)
				cleanup()
				c.Logger.Error(err, "cleanup ssh cache", "key", c)
			}()
			return nil, err
		}

		c.Logger.V(2).Info("wait ssh to close", "key", c)
		go func() {
			err := client.Wait()
			cleanup()
			c.Logger.Error(err, "wait ssh to close and cleanup", "key", c)
		}()

		return client, nil
	}
}
