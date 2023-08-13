package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(MatchOr{})
}

// MatchOr is ...
type MatchOr struct {
	MatchersRaw []json.RawMessage `json:"match" caddy:"namespace=dnsproxy.matchers inline_key=matcher"`

	matchers []Matcher
}

// CaddyModule is ...
func (MatchOr) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.matchers.or",
		New: func() caddy.Module { return new(MatchOr) },
	}
}

// Provision is ...
func (m *MatchOr) Provision(ctx caddy.Context) error {
	mods, err := ctx.LoadModule(m, "MatchersRaw")
	if err != nil {
		return err
	}
	for _, v := range mods.([]interface{}) {
		m.matchers = append(m.matchers, v.(Matcher))
	}
	return nil
}

// Match is ...
func (m *MatchOr) Match(in *dns.Msg) bool {
	for _, v := range m.matchers {
		if v.Match(in) {
			return true
		}
	}
	return false
}

var _ Matcher = (*MatchOr)(nil)
