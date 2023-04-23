package cache

import (
	"context"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/batx-dev/batproxy"
)

type ProxyService struct {
	next batproxy.ProxyService

	cache      *cache.Cache[string, *batproxy.Proxy]
	expiration time.Duration
}

type ProxyServiceOptions struct {
	ProxyExpiration time.Duration
}

func NewProxyService(next batproxy.ProxyService, opts ProxyServiceOptions) *ProxyService {
	s := &ProxyService{
		next:       next,
		cache:      cache.New[string, *batproxy.Proxy](),
		expiration: DefaultExpiration,
	}

	if opts.ProxyExpiration > 0 {
		s.expiration = opts.ProxyExpiration
	}

	return s
}

func (s *ProxyService) CreateProxy(ctx context.Context, proxy *batproxy.Proxy, opts batproxy.CreateProxyOptions) (err error) {
	defer func() {
		if err == nil {
			s.cache.Set(proxy.ID, proxy, cache.WithExpiration(s.expiration))
		}
	}()
	return s.next.CreateProxy(ctx, proxy, opts)
}

func (s *ProxyService) ListProxies(ctx context.Context, opts batproxy.ListProxiesOptions) (page *batproxy.ListProxiesPage, err error) {
	if opts.ProxyID != "" {
		res, ok := s.cache.Get(opts.ProxyID)
		if ok && res != nil {
			page = &batproxy.ListProxiesPage{
				Proxies: []*batproxy.Proxy{},
			}

			page.Proxies = append(page.Proxies, res)

			return page, nil
		}
	}

	page, err = s.next.ListProxies(ctx, opts)
	if err != nil {
		return nil, err
	}

	for _, p := range page.Proxies {
		s.cache.Set(p.ID, p, cache.WithExpiration(s.expiration))
	}

	return page, nil
}

func (s *ProxyService) DeleteProxy(ctx context.Context, proxyID string) (err error) {
	defer func() {
		if err == nil {
			s.cache.Delete(proxyID)
		}
	}()
	return s.next.DeleteProxy(ctx, proxyID)
}
