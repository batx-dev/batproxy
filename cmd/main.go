package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/batx-dev/batproxy/http"
	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/sql"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap/zapcore"
)

func main() {
	app := &cli.App{
		UseShortOptionHandling: true,
		Name:                   "batproxy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Usage:   "The listen address of batproxy",
				Value:   ":8888",
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

	l := logger.NewLogger(int8(2), "console", zapcore.ISO8601TimeEncoder)

	ll := l.Build().WithName("main")

	listen := cCtx.String("listen")

	dsn := cCtx.String("dsn")

	server, err := http.NewServer(listen, l)
	if err != nil {
		return err
	}

	db := sql.NewDB(dsn)
	if err := db.Open(); err != nil {
		return err
	}

	psvc := sql.NewProxy(db)

	server.ProxyService = psvc

	if err := server.Run(); err != nil {
		return err
	}

	ll.Info("main", "listen", listen)

	<-ctx.Done()

	return nil
}
