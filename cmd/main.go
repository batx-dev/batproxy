package main

import (
	"fmt"
	"log"
	"os"

	"github.com/batx-dev/batproxy"
	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Fprintf(cCtx.App.Writer, "barproxy version: %s\n", cCtx.App.Version)
	}

	app := cli.NewApp()

	app.Name = "batproxy"
	app.Commands = []*cli.Command{
		RunCmd(),
		ProxyCmd(),
	}
	app.Version = batproxy.Version

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
