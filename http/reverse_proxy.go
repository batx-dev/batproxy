package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/batx-dev/batproxy"
)

func (s *Server) reverseProxy(w http.ResponseWriter, req *http.Request) {
	reverseProxy, err := s.newReverseProxy(req)
	if err != nil {
		Error(w, req, err)
		return
	}
	reverseProxy.ServeHTTP(w, req)
}

func (s *Server) newReverseProxy(req *http.Request) (*httputil.ReverseProxy, error) {
	ctx := req.Context()

	proxyID := strings.Split(req.Host, ":")[0]

	ps, err := s.ProxyService.ListProxies(ctx, batproxy.ListProxiesOptions{
		ProxyID: proxyID,
	})
	if err != nil {
		return nil, err
	}

	if len(ps.Proxies) == 0 {
		return nil, fmt.Errorf("reverseProxy: can not find reverseProxy rule")
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

	sc, err := s.memo.Get(ctx, k)
	if err != nil {
		return nil, err
	}

	target = "http://" + target

	parse, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(parse)
	rp.Transport = &http.Transport{
		DialContext: sc.DialContext,
	}

	rp.ErrorHandler = s.reverseProxyHandlerError

	return rp, nil
}

func (s *Server) reverseProxyHandlerError(w http.ResponseWriter, req *http.Request, err error) {
	s.logger.Error(err, "reverse proxy", "req", req.Host)
	Error(w, req, err)
}
