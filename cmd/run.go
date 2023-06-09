package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/cache"
	"github.com/batx-dev/batproxy/http"
	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/sql"
	"github.com/urfave/cli/v2"
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
			&cli.StringFlag{
				Name:    "expiration",
				Usage:   "The time of proxy rule expiration",
				Value:   "15s",
				Aliases: []string{"e"},
				EnvVars: []string{"BATPROXY_EXPIRATION"},
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

	ll := logger.New(logger.Options{})

	reverseListen := cCtx.String("reverse-listen")
	listen := cCtx.String("listen")

	dsn := cCtx.String("dsn")

	suffix := cCtx.String("suffix")

	expiration := cCtx.String("expiration")

	server, err := http.NewServer(reverseListen, listen, ll.With("module", "http"))
	if err != nil {
		return err
	}

	db := sql.NewDB(dsn)
	if err := db.Open(); err != nil {
		return err
	}

	var psvc batproxy.ProxyService
	{
		duration, err := time.ParseDuration(expiration)
		if err != nil {
			return err
		}
		psvc = sql.NewProxyService(db, sql.ProxyServiceOptions{Suffix: suffix})
		psvc = cache.NewProxyService(psvc, cache.ProxyServiceOptions{ProxyExpiration: duration})
		psvc = logger.NewProxyService(psvc, ll.With("module", "logger"))
	}

	server.ProxyService = psvc

	if err := server.Open(); err != nil {
		return err
	}

	ll.Info("run", "module", "main", "reverse-listen", reverseListen)
	ll.Info("run", "module", "main", "listen", listen)
	ll.Info("run", "module", "main", "suffix", suffix)
	ll.Info("run", "module", "main", "expiration", expiration)

	<-ctx.Done()

	return server.Close()
}
