package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/proxy"
	"github.com/batx-dev/batproxy/ssh"
)

type Server struct {
	ln net.Listener

	Addr string

	transports map[string]func(ctx context.Context, network, address string) (net.Conn, error)

	batProxy *proxy.BatProxy
}

func NewServer(addr string, batProxy *proxy.BatProxy, logger logger.Logger) (*Server, error) {
	l := logger.Build().WithName("ssh")
	transports := make(map[string]func(ctx context.Context, network, address string) (net.Conn, error), 10)
	for _, p := range batProxy.Proxies {
		client := &ssh.Client{
			Host:         p.Host,
			User:         p.User,
			IdentityFile: p.IdentityFile,
			Password:     p.Password,
			Logger:       l,
		}

		s := &ssh.Ssh{
			Client: client,
		}

		transports[p.UUID] = s.DialContext

	}

	return &Server{
		Addr:       addr,
		transports: transports,
		batProxy:   batProxy,
	}, nil
}

func (s *Server) proxy(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	client := s.newClient(strings.Split(req.Host, ":")[0])

	request, err := client.newRequest(ctx, req.Method, req.URL.String(), req.Body)
	if err != nil {
		Error(w, req, err)
		return
	}
	request.Header = req.Header

	res, err := client.Do(request)
	if err != nil {
		Error(w, req, err)
		return
	}
	defer res.Body.Close()

	w.WriteHeader(res.StatusCode)

	content, err := io.ReadAll(res.Body)
	if err != nil {
		Error(w, req, err)
		return
	}

	if _, err := w.Write(content); err != nil {
		Error(w, req, err)
		return
	}
}

func (s *Server) Run() error {
	listen, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	http.HandleFunc("/", s.proxy)

	go http.Serve(listen, nil)

	return nil
}

func (s *Server) Stop() {
	err := s.ln.Close()
	if err != nil {
		return
	}
}
