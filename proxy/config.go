package proxy

import "fmt"

type BatProxy struct {
	Listen  string   `yaml:"listen"`
	Proxies []*Proxy `yaml:"proxies"`
}

type Proxy struct {
	UUID         string `yaml:"uuid"`
	User         string `yaml:"user"`
	Host         string `yaml:"host"`
	IdentityFile string `yaml:"identity_file,omitempty"`
	Password     string `yaml:"password,omitempty"`
	Node         string `yaml:"node"`
	Port         uint16 `yaml:"port"`
}

func (p *Proxy) Validate() error {
	if p.IdentityFile == "" && p.Password == "" {
		return fmt.Errorf("ssh: one of [passowrd, identity_file] required")
	}

	return nil
}

func NewBatProxy() BatProxy {
	return BatProxy{
		Proxies: []*Proxy{},
	}
}
