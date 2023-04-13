package ssh

import (
	"context"
	"time"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/memo"
	"golang.org/x/crypto/ssh"
)

func dialFunc(c *Client) memo.Func[key, *ssh.Client] {
	return func(ctx context.Context, key key, cleanup func()) (*ssh.Client, error) {
		if err := c.Validate(); err != nil {
			return nil, batproxy.Errorf(batproxy.EINVALID, "ssh client config: %s", err)
		}

		var auth []ssh.AuthMethod

		if c.PrivateKey != "" {
			var (
				err    error
				signer ssh.Signer
			)
			if c.Passphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(c.PrivateKey), []byte(c.Passphrase))
			} else {
				signer, err = ssh.ParsePrivateKey([]byte(c.PrivateKey))
			}
			if err != nil {
				return nil, batproxy.Errorf(batproxy.EINVALID, "parse private key: %v", err)
			} else {
				auth = append(auth, ssh.PublicKeys(signer))
			}
		}
		if c.Password != "" {
			auth = append(auth, ssh.Password(c.Password))
		}

		cfg := &ssh.ClientConfig{
			User:            c.User,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Second * 30,
		}

		// establish connect with remote host
		client, err := DialTimeout("tcp",
			c.Host,
			cfg,
			c.ServerAliveInterval*time.Duration(c.ServerAliveCountMax),
		)
		if err != nil {
			c.Logger.Error(err, "dial",
				"status", "fail",
				"key", key.String(),
			)
			go func() {
				time.Sleep(15 * time.Second)
				cleanup()
				c.Logger.Error(err, "cleanup ssh cache", "key", key.String())
			}()
			return nil, batproxy.Errorf(batproxy.EINTERNAL, "dial to %s", key.String())
		}

		c.Logger.V(2).Info("wait ssh to close", "key", key.String())

		go func() {
			err := client.Wait()
			cleanup()
			c.Logger.Error(err, "wait ssh to close and cleanup", "key", key.String())
		}()

		return client, nil
	}
}
