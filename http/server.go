package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/batx-dev/batproxy"
	"github.com/batx-dev/batproxy/logger"
	"github.com/batx-dev/batproxy/memo"
	"github.com/batx-dev/batproxy/ssh"
)

type Server struct {
	ln net.Listener

	memo *memo.Memo[key, *ssh.Ssh]

	Addr string

	ProxyService batproxy.ProxyService
}

func NewServer(addr string, logger logger.Logger) (*Server, error) {
	return &Server{
		Addr: addr,
		memo: memo.New(sshFunc(logger)),
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
	ctx := context.Background()
	uuid := strings.Split(req.Host, ":")[0]

	ps, err := s.ProxyService.ListProxies(ctx, batproxy.ListProxiesOptions{
		UUID: uuid,
	})
	if err != nil {
		return nil, err
	}

	if len(ps.Proxies) == 0 {
		return nil, fmt.Errorf("proxy: can not find proxy rule")
	}

	p := ps.Proxies[0]

	k := key{
		User:       p.User,
		Host:       p.Host,
		PrivateKey: p.PrivateKey,
		Passphrase: p.Passphrase,
		Password:   p.Password,
	}

	target := p.Node + ":" + strconv.Itoa(int(p.Port))

	sc, err := s.memo.Get(context.Background(), k)
	if err != nil {
		return nil, err
	}

	target = "http" + "://" + target
	parse, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(parse)
	rp.Transport = &http.Transport{
		DialContext: sc.DialContext,
	}
	return rp, nil
}

func (s *Server) Run() error {
	listen, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.ln = listen

	http.HandleFunc("/", s.proxy)

	go func() {
		err := http.Serve(listen, nil)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	err := s.ln.Close()
	if err != nil {
		return
	}
}
