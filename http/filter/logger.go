package filter

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/exp/slog"
)

// Logger recording http request.
func Logger(logger *slog.Logger) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		start := time.Now()

		chain.ProcessFilter(req, resp)

		logger.Info("",
			"method", req.Request.Method,
			"uri", req.Request.URL,
			"status", resp.StatusCode(),
			"size", resp.ContentLength(),
			"duration", time.Since(start),
			"ip", realIP(req.Request),
			"user_agent", req.HeaderParameter("User-Agent"),
		)
	}
}

var (
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

func realIP(r *http.Request) string {
	if xrip := r.Header.Get(xRealIP); xrip != "" {
		return xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		return xff[:i]
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		} else {
			return host
		}
	}
}
