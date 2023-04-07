package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Client represents an HTTP client.
type Client struct {
	client http.Client

	URL string
}

func (s *Server) newClient(uuid string) *Client {
	dcf, ok := s.transports[uuid]
	if !ok {
		panic(uuid)
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: dcf,
		},
	}

	url := ""
	for _, p := range s.batProxy.Proxies {
		if uuid == p.UUID {
			url = p.Node + ":" + strconv.Itoa(int(p.Port))
			break
		}

	}
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	return &Client{URL: url, client: client}
}

// newRequest returns a new HTTP request.
func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	// Build new request with base URL.
	req, err := http.NewRequest(method, c.URL+url, body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, req *http.Request, err error) {
	fmt.Fprintf(w, "%s", err)
}
