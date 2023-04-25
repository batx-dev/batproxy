package logger

import (
	"context"
	"time"

	"github.com/batx-dev/batproxy"
	"golang.org/x/exp/slog"
)

type ProxyService struct {
	logger *slog.Logger
	next   batproxy.ProxyService
}

func NewProxyService(next batproxy.ProxyService, logger *slog.Logger) batproxy.ProxyService {
	return &ProxyService{
		logger: logger,
		next:   next,
	}
}

func (s *ProxyService) CreateProxy(ctx context.Context, proxy *batproxy.Proxy, opts batproxy.CreateProxyOptions) (err error) {
	defer func(begin time.Time) {
		logger := s.logger.With(
			"took", time.Since(begin),
			"proxy_id", proxy.ID,
			"user", proxy.User,
			"host", proxy.Host,
			"node", proxy.Node,
			"port", proxy.Port,
		)
		logErr(logger, "CreateProxy", err)
	}(time.Now())
	return s.next.CreateProxy(ctx, proxy, opts)
}

func (s *ProxyService) ListProxies(ctx context.Context, opts batproxy.ListProxiesOptions) (page *batproxy.ListProxiesPage, err error) {
	defer func(begin time.Time) {
		logger := s.logger.With(
			"took", time.Since(begin),
			"proxy_id", opts.ProxyID,
			"page_token", opts.PageToken,
			"page_size", opts.PageSize,
			"num", func() int {
				if page != nil {
					return len(page.Proxies)
				}
				return 0
			}(),
		)
		logErr(logger, "ListProxies", err)
	}(time.Now())
	return s.next.ListProxies(ctx, opts)
}

func (s *ProxyService) DeleteProxy(ctx context.Context, proxyID string) (err error) {
	defer func(begin time.Time) {
		logger := s.logger.With(
			"took", time.Since(begin),
			"proxy_id", proxyID,
		)
		logErr(logger, "DeleteProxy", err)
	}(time.Now())
	return s.next.DeleteProxy(ctx, proxyID)
}
