package main

import (
	"fmt"
	"strings"

	"github.com/batx-dev/batproxy"
	"github.com/urfave/cli/v2"
)

func ProxyCmd() *cli.Command {
	cmd := &cli.Command{
		Name:  "proxy",
		Usage: "Manage proxy rule",
		Subcommands: []*cli.Command{
			ProxyCreateCmd(),
			ProxiesListCmd(),
			ProxyDeleteCmd(),
		},
	}

	return cmd
}

func unixSocketFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "base-url",
		Usage:   "The manager proxy listener address",
		Value:   "unix://batproxy.sock",
		Aliases: []string{"l"},
		EnvVars: []string{"BATPROXY_LISTEN", "BATPROXY_BASE_URL"},
		Action: func(c *cli.Context, s string) error {
			ss := strings.Split(s, "://")

			if len(ss) < 2 {
				if err := c.Set("base-url", fmt.Sprintf("unix://%s", s)); err != nil {
					return err
				}

				return nil
			}

			switch ss[0] {
			case "unix", "http", "https":
				return nil
			case "tcp", "udp":
				if err := c.Set("base-url", fmt.Sprintf("http://%s", ss[1])); err != nil {
					return err
				}
				return nil
			default:
				return batproxy.Errorf(batproxy.EINVALID, "expect scheme ['unix', 'http', 'https'], got %s", ss[0])
			}
		},
	}
}
