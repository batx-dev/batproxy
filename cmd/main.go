package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/batx-dev/batproxy/http"
	"github.com/batx-dev/batproxy/sql"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

func main() {
	app := &cli.App{
		UseShortOptionHandling: true,
		Name:                   "batproxy",
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
				Value:   ":18888",
				Aliases: []string{"l"},
				EnvVars: []string{"BATPROXY_LISTEN"},
			},
			&cli.StringFlag{
				Name:    "dsn",
				Usage:   "The database dsn ( file path if sqlite3 )",
				Value:   "batproxy.db",
				Aliases: []string{"d"},
				EnvVars: []string{"BATPROXY_DSN"},
			},
		},
		Action: Run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Run(cCtx *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
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

	server, err := http.NewServer(rl, l, logger)
	if err != nil {
		return err
	}

	db := sql.NewDB(dsn)
	if err := db.Open(); err != nil {
		return err
	}

	psvc := sql.NewProxy(db)

	server.ProxyService = psvc

	if err := server.Open(); err != nil {
		return err
	}

	logger.Info("main", "reverse-listen", rl)
	logger.Info("main", "listen", l)

	<-ctx.Done()

	return nil
}
