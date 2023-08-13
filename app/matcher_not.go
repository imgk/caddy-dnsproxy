package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(MatchNot{})
}

// MatchNot is ...
type MatchNot struct {
	MatcherRaw json.RawMessage `json:"match" caddy:"namespace=dnsproxy.matchers inline_key=matcher"`

	matcher Matcher
}

// CaddyModule is ...
func (MatchNot) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.matchers.not",
		New: func() caddy.Module { return new(MatchNot) },
	}
}

// Provision is ...
func (m *MatchNot) Provision(ctx caddy.Context) error {
	mod, err := ctx.LoadModule(m, "MatcherRaw")
	if err != nil {
		return err
	}
	m.matcher = mod.(Matcher)
	return nil
}

// Match is ...
func (*MatchNot) Match(_ *dns.Msg) bool {
	return true
}

var _ Matcher = (*MatchNot)(nil)
