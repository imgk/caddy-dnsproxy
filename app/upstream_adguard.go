package app

import (
	"github.com/caddyserver/caddy/v2"

	"github.com/AdguardTeam/dnsproxy/upstream"
)

func init() {
	caddy.RegisterModule(AdGuard{})
}

// AdGuard is ...
type AdGuard struct {
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
func (AdGuard) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.upstreams.adguard",
		New: func() caddy.Module { return new(AdGuard) },
	}
}

// Provision is ...
func (m *AdGuard) Provision(ctx caddy.Context) error {
	up, err := upstream.AddressToUpstream(m.Server, nil)
	if err != nil {
		return err
	}
	m.Upstream = up
	return nil
}

var _ Upstream = (*AdGuard)(nil)
