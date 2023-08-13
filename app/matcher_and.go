package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(MatchAnd{})
}

// MatchAnd is ...
type MatchAnd struct {
	MatchersRaw []json.RawMessage `json:"match" caddy:"namespace=dnsproxy.matchers inline_key=matcher"`

	matchers []Matcher
}

// CaddyModule is ...
func (MatchAnd) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.matchers.and",
		New: func() caddy.Module { return new(MatchAnd) },
	}
}

// Provision is ...
func (m *MatchAnd) Provision(ctx caddy.Context) error {
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
func (m *MatchAnd) Match(in *dns.Msg) bool {
	for _, v := range m.matchers {
		if !v.Match(in) {
			return false
		}
	}
	return true
}

var _ Matcher = (*MatchAnd)(nil)
