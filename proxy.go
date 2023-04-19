package batproxy

import (
	"context"
	"fmt"
	"time"
)

type Proxy struct {
	// ID Unique proxy id.
	// Format: <uuid><.suffix>
	// Output only.
	ID string `json:"proxy_id"`

	// User Over SSH login name.
	// Required.
	User string `json:"user"`

	// Host Over SSH login host.
	// Required.
	Host string `json:"host"`

	// PrivateKey Over SSH login private key.
	// Optional.
	PrivateKey string `json:"private_key,omitempty"`

	// Passphrase Over SSH login private key passphrase.
	// Optional.
	Passphrase string `json:"passphrase,omitempty"`

	// Password Over SSH login password.
	// Optional.
	Password string `json:"password,omitempty"`

	// Node Proxy to destination.
	// Required.
	Node string `json:"node"`

	// Port Proxy to destination.
	// Required.
	Port uint16 `json:"port"`

	// CreateTime Create time of this address.
	// Output only.
	CreateTime time.Time `json:"create_time"`

	// UpdateTime Update time of this address.
	// Output only.
	UpdateTime time.Time `json:"update_time"`
}

func (p *Proxy) Validate() error {
	if p.User == "" || p.Host == "" {
		return fmt.Errorf("invalid ssh format user@host: %s@%s", p.User, p.Host)
	}

	if p.PrivateKey == "" && p.Password == "" {
		return fmt.Errorf("ssh auth required one of [passowrd, private_key]")
	}

	if p.Node == "" || p.Port == 0 {
		return fmt.Errorf("invalid proxy destination %s:%d", p.Node, p.Port)
	}

	return nil
}

type CreateProxyOptions struct {
	// Suffix will append after uuid
	// Format: <uuid><.suffix>
	// Optional.
	Suffix string `schema:"suffix,omitempty"`
}

type ListProxiesPage struct {
	Proxies       []*Proxy `json:"proxies" schema:"proxies"`
	NextPageToken string   `json:"next_page_token,omitempty" schema:"next_page_token,omitempty"`
}

type ListProxiesOptions struct {
	// ProxyID unique proxy rule id.
	ProxyID string `schema:"proxy_id,omitempty"`

	// PageSize sets the maximum number of users to be returned.
	// 0 means no maximum; driver implementations should choose a reasonable
	// max. It is guaranteed to be >= 0.
	PageSize int `schema:"page_size,omitempty"`

	// PageToken may be filled in with the NextPageToken from a previous
	// ListUsers call.
	PageToken string `schema:"page_token,omitempty"`
}

type ProxyService interface {
	CreateProxy(ctx context.Context, proxy *Proxy, opts CreateProxyOptions) error
	ListProxies(ctx context.Context, opts ListProxiesOptions) (*ListProxiesPage, error)
	DeleteProxy(ctx context.Context, proxyID string) error
}
