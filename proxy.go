package batproxy

import (
	"context"
	"fmt"
	"time"
)

type Proxy struct {
	// UUID Unique proxy id.
	// Output only.
	UUID string `json:"uuid"`

	// User Over SSH login name.
	// Required.
	User string `json:"user"`

	// Host Over SSH login host.
	// Required.
	Host string `json:"host"`

	// PrivateKey Over SSH login private key.
	// Optional.
	PrivateKey string `json:"private_key,omitempty"`

	// Passphrase Over SSH login private key password.
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
	if p.PrivateKey == "" && p.Password == "" {
		return fmt.Errorf("ssh auth required one of [passowrd, private_key]")
	}

	return nil
}

func (p *Proxy) String() string {
	fmt.Printf("111: %+v\n", 111)
	return fmt.Sprintf("%s@%s", p.User, p.Host)
}

func (p *Proxy) Equal(other *Proxy) int {
	if p.Host == other.Host && p.User == other.User {
		return 0
	}
	fmt.Printf("p: %+v\n", p)
	fmt.Printf("other: %+v\n", other)

	return -1
}

type ListProxiesPage struct {
	Proxies       []*Proxy `json:"proxies" schema:"proxies"`
	NextPageToken string   `json:"next_page_token,omitempty" schema:"next_page_token,omitempty"`
}

type ListProxiesOptions struct {
	UUID string
	// PageSize sets the maximum number of users to be returned.
	// 0 means no maximum; driver implementations should choose a reasonable
	// max. It is guaranteed to be >= 0.
	PageSize int `schema:"page_size,omitempty"`
	// PageToken may be filled in with the NextPageToken from a previous
	// ListUsers call.
	PageToken string `schema:"page_token,omitempty"`
}

type ProxyService interface {
	ListProxies(ctx context.Context, opts ListProxiesOptions) (*ListProxiesPage, error)
}
