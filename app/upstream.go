package app

import (
	"github.com/caddyserver/caddy/v2"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(Adguard{})
}

// Upstream is ...
type Upstream interface {
	Exchange(m *dns.Msg) (*dns.Msg, error)
}

// Adguard is ...
type Adguard struct {
	// Server is ...
	Server string `json:"server"`
	// Bootstrap is ...
	Bootstrap string `json:"bootstrap,omitempty"`
	// Timeout is ...
	Timeout caddy.Duration `json:"timeout,omitempty"`
	// Upstream is ...
	upstream.Upstream `json:"-,omitempty"`
}

// CaddyModule is ...
func (Adguard) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.upstreams.adguard",
		New: func() caddy.Module { return new(Adguard) },
	}
}

// Provision is ...
func (m *Adguard) Provision(ctx caddy.Context) error {
	up, err := upstream.AddressToUpstream(m.Server, nil)
	if err != nil {
		return err
	}
	m.Upstream = up
	return nil
}

var _ Upstream = (*Adguard)(nil)
