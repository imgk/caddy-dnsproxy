package app

import (
	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"

	"github.com/imgk/caddy-dnsproxy/pkg/suffixtree"
)

func init() {
	caddy.RegisterModule(MatchDomain{})
}

// MatchDomain is ...
type MatchDomain struct {
	// Domains is ...
	Domains []string `json:"domains,omitempty"`

	node *suffixtree.Node
}

// CaddyModule is ...
func (MatchDomain) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.matchers.match_domain",
		New: func() caddy.Module { return new(MatchDomain) },
	}
}

// Provision is ...
func (m *MatchDomain) Provision(ctx caddy.Context) error {
	m.node = suffixtree.NewNodeFromRules(m.Domains...)
	return nil
}

// Match is ...
func (m *MatchDomain) Match(in *dns.Msg) bool {
	for _, v := range in.Question {
		if m.node.Match(v.Name) {
			return true
		}
	}
	return false
}

var (
	_ caddy.Provisioner = (*MatchDomain)(nil)
	_ Matcher           = (*MatchDomain)(nil)
)
