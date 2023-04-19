package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/http"
	"github.com/batx-dev/batproxy/sql"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

func RunCmd() *cli.Command {
	cmd := &cli.Command{
		Name:                   "run",
		UseShortOptionHandling: true,
		Usage:                  "run batproxy service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "reverse-listen",
				Usage:   "The reverse proxy http server listen address",
				Value:   ":8888",
				Aliases: []string{"r"},
				EnvVars: []string{"BATPROXY_REVERSE_LISTEN"},
			},
			&cli.StringFlag{
				Name:    "listen",
				Usage:   "The manager proxy listen address",
				Value:   "unix://batproxy.sock",
				Aliases: []string{"l"},
				EnvVars: []string{"BATPROXY_LISTEN"},
				Action: func(c *cli.Context, s string) error {
					ss := strings.Split(s, "://")

					if len(ss) < 2 {
						if err := c.Set("listen", fmt.Sprintf("unix://%s", s)); err != nil {
							return err
						}

						return nil
					}

					switch ss[0] {
					case "unix", "tcp", "udp":
						return nil
					default:
						return batproxy.Errorf(batproxy.EINVALID, "network: %s", ss[0])
					}

				},
			},
			&cli.StringFlag{
				Name:    "suffix",
				Usage:   "The proxy id default suffix",
				Aliases: []string{"s"},
				EnvVars: []string{"BATPROXY_PROXY_SUFFIX"},
			},
			&cli.StringFlag{
				Name:    "dsn",
				Usage:   "The database dsn ( file path if sqlite3 )",
				Value:   "batproxy.db",
				Aliases: []string{"d"},
				EnvVars: []string{"BATPROXY_DSN"},
			},
		},
		Action: RunAction,
	}

	return cmd
}

func RunAction(cCtx *cli.Context) error {
	ctx, cancel := context.WithCancel(cCtx.Context)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	loggerOption := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(loggerOption.NewTextHandler(os.Stdout))

	rl := cCtx.String("reverse-listen")
	l := cCtx.String("listen")

	dsn := cCtx.String("dsn")
	suffix := cCtx.String("suffix")

	server, err := http.NewServer(rl, l, logger.With("msg", "http"))
	if err != nil {
		return err
	}

	db := sql.NewDB(dsn)
	if err := db.Open(); err != nil {
		return err
	}

	psvc := sql.NewProxy(db, sql.ProxyServiceOptions{Suffix: suffix})

	server.ProxyService = psvc

	if err := server.Open(); err != nil {
		return err
	}

	logger.Info("main", "reverse-listen", rl)
	logger.Info("main", "listen", l)

	<-ctx.Done()

	return server.Close()
}
