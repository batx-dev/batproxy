package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/batx-dev/batproxy/http"
	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/proxy"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	l := logger.NewLogger(int8(2), "console", zapcore.ISO8601TimeEncoder)

	batProxy := &proxy.BatProxy{Proxies: []*proxy.Proxy{}}
	if len(os.Args) > 1 {
		data, err := os.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		if err := yaml.Unmarshal(data, batProxy); err != nil {
			panic(err)
		}
	}

	server, err := http.NewServer(batProxy.Listen, batProxy, l)
	if err != nil {
		panic(err)
	}

	if err := server.Run(); err != nil {
		panic(err)
	}

	<-ctx.Done()
}
