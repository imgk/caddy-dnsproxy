package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(Cache{})
}

// Cache is ...
type Cache struct {
	UpstreamRaw json.RawMessage `json:"upstream" caddy:"namespace=dnsproxy.upstreams inline_key=upstream"`

	upstream Upstream
	data     map[string]dns.RR
}

// CaddyModule is ...
func (Cache) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.upstreams.Cache",
		New: func() caddy.Module { return new(Cache) },
	}
}

// Provision is ...
func (m *Cache) Provision(ctx caddy.Context) error {
	m.data = map[string]dns.RR{}
	mod, err := ctx.LoadModule(m, "UpstreamRaw")
	if err != nil {
		return err
	}
	m.upstream = mod.(Upstream)
	return nil
}

// Exchange is ...
func (m *Cache) Exchange(in *dns.Msg) (*dns.Msg, error) {
	for i := range in.Question {
		rr, ok := m.data[in.Question[i].Name]
		if ok {
			in.MsgHdr.Response = true
			in.Answer = append(in.Answer[:0], rr)
			return in, nil
		}
	}
	out, err := m.upstream.Exchange(in)
	if err != nil || len(out.Answer) == 0 {
		return out, err
	}
	m.data[in.Question[0].Name] = out.Answer[0]
	return out, nil
}

var _ Upstream = (*Cache)(nil)
