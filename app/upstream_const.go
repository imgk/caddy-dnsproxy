package app

import (
	"errors"
	"net"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterModule(Const{})
}

// Const is ...
type Const struct {
	// Type is ...
	Type string `json:"type,omitempty"`
	// Name is ...
	Name string `json:"name"`

	qType  uint16
	qValue net.IP
}

// CaddyModule is ...
func (Const) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dnsproxy.upstreams.const",
		New: func() caddy.Module { return new(Const) },
	}
}

// Provision is ...
func (m *Const) Provision(ctx caddy.Context) error {
	if t, ok := dns.StringToType[m.Type]; !ok {
		return errors.New("invalid type")
	} else {
		m.qType = t
	}

	if ip := net.ParseIP(m.Name); ip == nil {
		return errors.New("")
	} else {
		m.qValue = ip
	}

	return nil
}

// Exchange is ...
func (m *Const) Exchange(in *dns.Msg) (*dns.Msg, error) {
	in.MsgHdr.Response = true
	for _, v := range in.Question {
		if v.Qtype == m.qType {
			switch m.qType {
			case dns.TypeA:
				in.Answer = append(in.Answer[:0], &dns.A{
					Hdr: dns.RR_Header{},
					A:   m.qValue,
				})
			case dns.TypeAAAA:
				in.Answer = append(in.Answer[:0], &dns.AAAA{
					Hdr:  dns.RR_Header{},
					AAAA: m.qValue,
				})
			default:
			}
			return in, nil
		}
	}
	return in, nil
}

var _ Upstream = (*Const)(nil)
