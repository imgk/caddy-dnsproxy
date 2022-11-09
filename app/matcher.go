package app

import "github.com/miekg/dns"

// Matcher is ...
type Matcher interface {
	// Match is ...
	Match(*dns.Msg) bool
}
