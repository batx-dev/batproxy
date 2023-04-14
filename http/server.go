package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/memo"
	"github.com/batx-dev/batproxy/ssh"
	"github.com/emicklei/go-restful/v3"
	"golang.org/x/exp/slog"
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

type Server struct {
	memo *memo.Memo[key, *ssh.Ssh]

	logger *slog.Logger

	managerListen net.Listener
	managerServer *http.Server
	managerAddr   string

	reverseProxyListen net.Listener
	reverseProxyServer *http.Server
	reverseProxyAddr   string

	ProxyService batproxy.ProxyService
}

func NewServer(reverseProxyAddr, managerAddr string, logger *slog.Logger) (*Server, error) {
	return &Server{
		memo:             memo.New(sshFunc(logger)),
		logger:           logger,
		managerAddr:      managerAddr,
		reverseProxyAddr: reverseProxyAddr,
	}, nil
}

func (s *Server) Open() (err error) {
	// listen reverse reverseProxy address
	{
		s.reverseProxyServer = &http.Server{}
		handleFunc := http.HandlerFunc(s.reverseProxy)
		s.reverseProxyServer.Handler = handleFunc

		if s.reverseProxyListen, err = net.Listen("tcp", s.reverseProxyAddr); err != nil {
			return err
		}

		go func() {
			if err := s.reverseProxyServer.Serve(s.reverseProxyListen); err != nil {
				panic(err)
			}
		}()
	}

	// listen manager reverseProxy address
	{
		c := restful.NewContainer()
		corev1beta1 := new(restful.WebService)
		corev1beta1.Path("/api/v1beta1").
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON)

		s.proxyService(corev1beta1)

		c.Add(corev1beta1)

		s.managerServer = &http.Server{}
		s.managerServer.Handler = c

		if s.managerListen, err = net.Listen("tcp", s.managerAddr); err != nil {
			return err
		}

		go func() {
			err := s.managerServer.Serve(s.managerListen)
			if err != nil {
				panic(err)
			}
		}()
	}

	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	if err := s.reverseProxyServer.Shutdown(ctx); err != nil {
		return err
	}

	if err := s.managerServer.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
