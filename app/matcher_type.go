package app

import (
	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(MatchType{})
}

// MatchType is ...
type MatchType struct {
	Types []string `json:"query_type"`

	typeList []uint16
}

// CaddyModule is ...
func (MatchType) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.matchers.query_type",
		New: func() caddy.Module { return new(MatchType) },
	}
}

// Provision is ...
func (m *MatchType) Provision(ctx caddy.Context) error {
	for _, v := range m.Types {
		m.typeList = append(m.typeList, dns.StringToType[v])
	}
	return nil
}

// Match is ...
func (m *MatchType) Match(in *dns.Msg) bool {
	for i := range in.Question {
		for _, v := range m.typeList {
			if in.Question[i].Qtype == v {
				return true
			}
		}
	}
	return false
}

var _ Matcher = (*MatchType)(nil)
