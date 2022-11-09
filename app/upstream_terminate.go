package app

import (
	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(Terminate{})
}

// Terminate is ...
type Terminate struct{}

// CaddyModule is ...
func (Terminate) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.upstreams.terminate",
		New: func() caddy.Module { return new(Terminate) },
	}
}

// Exchange is ...
func (m *Terminate) Exchange(in *dns.Msg) (*dns.Msg, error) {
	return in, nil
}

var _ Upstream = (*Terminate)(nil)
