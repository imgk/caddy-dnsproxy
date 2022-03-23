package app

import (
	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(MatchAll{})
}

// Matcher is ...
type Matcher interface {
	// Match is ...
	Match(*dns.Msg) bool
}

// MatchAll is ...
type MatchAll struct{}

// CaddyModule is ...
func (MatchAll) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.matchers.match_all",
		New: func() caddy.Module { return new(MatchAll) },
	}
}

// Match is ...
func (*MatchAll) Match(_ *dns.Msg) bool {
	return true
}

var _ Matcher = (*MatchAll)(nil)
