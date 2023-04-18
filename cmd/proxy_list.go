package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/http"
	"github.com/urfave/cli/v2"
)

func ProxiesListCmd() *cli.Command {
	cmd := &cli.Command{
		Name:  "list",
		Usage: "list proxies rule",
		Flags: []cli.Flag{
			unixSocketFlag(),
			&cli.StringFlag{
				Name:    "name",
				Usage:   "Proxy id",
				Aliases: []string{"n"},
			},
			&cli.StringFlag{
				Name:  "node",
				Usage: "Proxy to destination",
			},
			&cli.UintFlag{
				Name:  "port",
				Usage: "Proxy to destination",
			},
		},

		Action: ProxiesListAction,
	}
	return cmd
}

func ProxiesListAction(cCtx *cli.Context) error {
	opts := batproxy.ListProxiesOptions{
		ProxyID:  cCtx.String("name"),
		PageSize: 30,
	}

	client, err := http.NewClient(cCtx.String("base-url"))
	if err != nil {
		return err
	}

	svc := http.ProxyService{
		Client: client,
	}

	page, err := svc.ListProxies(cCtx.Context, opts)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "NAME\tUSER\tHOST\tNODE\tPORT\n")
	for _, p := range page.Proxies {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n", p.ID, p.User, p.Host, p.Node, p.Port)
	}

	return tw.Flush()
}
