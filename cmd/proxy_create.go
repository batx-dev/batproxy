package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/http"
	"github.com/urfave/cli/v2"
)

func ProxyCreateCmd() *cli.Command {
	cmd := &cli.Command{
		Name:  "create",
		Usage: "create proxy rule",
		Flags: []cli.Flag{
			unixSocketFlag(),
			&cli.StringFlag{
				Name:     "suffix",
				Usage:    "Proxy id suffix",
				Category: "PROXY",
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Proxy id, will overlay <suffix>",
				Aliases:  []string{"n"},
				Category: "PROXY",
			},
			&cli.StringFlag{
				Name:     "user",
				Usage:    "Over SSH login name",
				Aliases:  []string{"u"},
				Required: true,
				Category: "SSH",
			},
			&cli.StringFlag{
				Name:     "host",
				Usage:    "Over SSH login host, contains port",
				Aliases:  []string{"H"},
				Required: true,
				Category: "SSH",
			},
			&cli.StringFlag{
				Name:     "private-key",
				Usage:    "Over SSH login private key",
				Aliases:  []string{"i"},
				Category: "SSH",
			},
			&cli.StringFlag{
				Name:     "passphrase",
				Usage:    "Over SSH login private key passphrase",
				Aliases:  []string{"s"},
				Category: "SSH",
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "Over SSH login password",
				Aliases:  []string{"p"},
				Category: "SSH",
			},
			&cli.StringFlag{
				Name:     "node",
				Usage:    "Proxy to destination",
				Required: true,
				Category: "PROXY",
			},
			&cli.UintFlag{
				Name:     "port",
				Usage:    "Proxy to destination",
				Required: true,
				Category: "PROXY",
			},
		},
		Action: ProxyCreateAction,
	}

	return cmd
}

func ProxyCreateAction(cCtx *cli.Context) error {
	proxy := &batproxy.Proxy{
		ID:         cCtx.String("name"),
		User:       cCtx.String("user"),
		Host:       cCtx.String("host"),
		PrivateKey: cCtx.String("private-key"),
		Passphrase: cCtx.String("passphrase"),
		Password:   cCtx.String("password"),
		Node:       cCtx.String("node"),
		Port:       uint16(cCtx.Uint("port")),
	}
	if err := proxy.Validate(); err != nil {
		return err
	}

	opts := batproxy.CreateProxyOptions{Suffix: cCtx.String("suffix")}

	client, err := http.NewClient(cCtx.String("base-url"))
	if err != nil {
		return err
	}

	svc := http.ProxyService{
		Client: client,
	}

	if err := svc.CreateProxy(cCtx.Context, proxy, opts); err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "NAME\tUSER\tHOST\tNODE\tPORT\n")
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n", proxy.ID, proxy.User, proxy.Host, proxy.Node, proxy.Port)

	return tw.Flush()
}
