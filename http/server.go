package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
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
	reverseProxy, err := s.NewProxy(req)
	if err != nil {
		Error(w, req, err)
		return
	}
	reverseProxy.ServeHTTP(w, req)
}

func (s *Server) NewProxy(req *http.Request) (*httputil.ReverseProxy, error) {
	uuid := strings.Split(req.Host, ":")[0]
	dcf, ok := s.transports[uuid]
	if !ok {
		panic(uuid)
	}
	target := ""
	for _, p := range s.batProxy.Proxies {
		if uuid == p.UUID {
			target = p.Node + ":" + strconv.Itoa(int(p.Port))
			break
		}

	}

	target = "http" + "://" + target
	parse, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(parse)
	rp.Transport = &http.Transport{
		DialContext: dcf,
	}
	return rp, nil
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
