package app

import "github.com/miekg/dns"

// Upstream is ...
type Upstream interface {
	// Exchange is ...
	Exchange(m *dns.Msg) (*dns.Msg, error)
}
