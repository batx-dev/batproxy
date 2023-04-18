package main

import (
	"fmt"

	"github.com/batx-dev/batproxy/http"
	"github.com/urfave/cli/v2"
)

func ProxyDeleteCmd() *cli.Command {
	cmd := &cli.Command{
		Name:  "delete",
		Usage: "list proxies rule",
		Flags: []cli.Flag{
			unixSocketFlag(),
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Proxy id",
				Aliases:  []string{"n"},
				Required: true,
			},
		},

		Action: ProxyDeleteAction,
	}

	return cmd
}

func ProxyDeleteAction(cCtx *cli.Context) error {
	client, err := http.NewClient(cCtx.String("base-url"))
	if err != nil {
		return err
	}

	svc := http.ProxyService{
		Client: client,
	}
	proxyID := cCtx.String("name")
	if err := svc.DeleteProxy(cCtx.Context, proxyID); err != nil {
		return err
	}

	fmt.Printf("Deleted: %s\n", proxyID)

	return nil
}
