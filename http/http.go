package http

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/batx-dev/batproxy"
	"github.com/gorilla/schema"
)

var (
	decoder = schema.NewDecoder()
	encoder = schema.NewEncoder()
)

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, req *http.Request, err error) {
	code, message := batproxy.ErrorCode(err), batproxy.ErrorMessage(err)
	w.WriteHeader(ErrorStatusCode(code))
	_, _ = w.Write([]byte(message))
}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error string `json:"error"`
}

// parseResponseError parses an JSON-formatted error response.
func parseResponseError(res *http.Response) error {
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var errRes ErrorResponse
	if err := json.Unmarshal(buf, &errRes); err != nil {
		message := strings.TrimSpace(string(buf))
		if message == "" {
			message = "Empty response from server."
		}
		return batproxy.Errorf(FromErrorStatusCode(res.StatusCode), message)
	}
	return batproxy.Errorf(FromErrorStatusCode(res.StatusCode), errRes.Error)
}

// lookup of application error codes to HTTP status codes.
var codes = map[string]int{
	batproxy.ECONFLICT:       http.StatusConflict,
	batproxy.EINVALID:        http.StatusBadRequest,
	batproxy.ENOTFOUND:       http.StatusNotFound,
	batproxy.ENOTIMPLEMENTED: http.StatusNotImplemented,
	batproxy.EUNAUTHORIZED:   http.StatusUnauthorized,
	batproxy.EINTERNAL:       http.StatusInternalServerError,
	batproxy.EFORBIDDEN:      http.StatusForbidden,
	batproxy.EBADGATEWAY:     http.StatusBadGateway,
}

// ErrorStatusCode returns the associated HTTP status code for a BatProxy error code.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// FromErrorStatusCode returns the associated BatProxy code for an HTTP status code.
func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return batproxy.EINTERNAL
}

// Client represents an HTTP client.
type Client struct {
	client *http.Client

	URL string
}

// NewClient returns a new instance of Client.
func NewClient(u string) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,
	}

	ss := strings.Split(u, "://")
	if len(ss) < 2 {
		return nil, batproxy.Errorf(batproxy.EINVALID, "base url %s", u)
	}

	switch ss[0] {
	case "unix":
		c.client.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return net.Dial(ss[0], ss[1])
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		c.URL = "http://localhost"
	case "http", "https":
		c.URL = u
	default:
		return nil, batproxy.Errorf(batproxy.EINVALID, "expect scheme ['unix', 'http', 'https'], got %s", ss[0])
	}

	return c, nil
}

// newRequest returns a new HTTP request.
func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	// Build new request with base URL.
	req, err := http.NewRequest(method, c.URL+url, body)
	if err != nil {
		return nil, err
	}

	// Default to JSON format.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// AllToEmpty convert `-` which means all to empty.
func AllToEmpty(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

// EmptyToAll covert empty to `-` which mean all.
func EmptyToAll(s string) string {
	if s == "" {
		return "-"
	}
	return s
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
