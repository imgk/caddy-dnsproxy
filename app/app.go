package app

import (
	"encoding/json"
	"errors"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const CaddyAppID = "dnsproxy"

const (
	DefaultUDPPort  = 53
	DefaultTCPPort  = 53
	DefaultTLSPort  = 853
	DefaultQuicPort = 853
)

func init() {
	caddy.RegisterModule(App{})
}

// App is ...
type App struct {
	// Handlers is ...
	Handlers []*struct {
		UpstreamRaw json.RawMessage   `json:"upstream" caddy:"namespace=dnsproxy.upstreams inline_key=upstream"`
		MatchersRaw []json.RawMessage `json:"match" caddy:"namespace=dnsproxy.matchers inline_key=matcher"`
	} `json:"handlers"`
	// ListenUDP is ...
	ListenUDP int `json:"udp,omitempty"`
	// ListenTCP is ...
	ListenTCP int `json:"tcp,omitempty"`
	// ListenTLS is ...
	ListenTLS int `json:"tls,omitempty"`
	// ListenQuic is ...
	ListenQuic int `json:"quic,omitempty"`
	// Servers is ...
	Servers []string `json:"servers,omitempty"`

	lg       *zap.Logger
	handlers []Handler
}

// CaddyModule is ...
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  CaddyAppID,
		New: func() caddy.Module { return new(App) },
	}
}

// Provision is ...
func (app *App) Provision(ctx caddy.Context) error {
	if app.ListenUDP == 0 {
		app.ListenUDP = DefaultUDPPort
	}
	if app.ListenTCP == 0 {
		app.ListenTCP = DefaultTCPPort
	}
	if app.ListenTLS == 0 {
		app.ListenTLS = DefaultTLSPort
	}
	if app.ListenQuic == 0 {
		app.ListenQuic = DefaultQuicPort
	}

	app.lg = ctx.Logger(app)

	for _, v := range app.Handlers {
		hd := Handler{}

		// parse upstream
		mod, err := ctx.LoadModule(v, "UpstreamRaw")
		if err != nil {
			return err
		}
		hd.Upstream = mod.(Upstream)

		// parse matchers
		mods, err := ctx.LoadModule(v, "MatchersRaw")
		if err != nil {
			return err
		}
		for _, v := range mods.([]interface{}) {
			hd.Matchers = append(hd.Matchers, v.(Matcher))
		}

		app.handlers = append(app.handlers, hd)
	}

	return nil
}

// Validate is ...
func (app *App) Validate() error {
	return nil
}

// Start is ...
func (app *App) Start() error {
	return nil
}

// Stop is ...
func (app *App) Stop() error {
	return nil
}

// Cleanup is ...
func (app *App) Cleanup() error {
	for _, v := range app.handlers {
		for _, vv := range v.Matchers {
			if cu, ok := vv.(caddy.CleanerUpper); ok {
				cu.Cleanup()
			}
		}
	}
	return nil
}

// Exchange is ...
func (app *App) Exchange(in *dns.Msg) (*dns.Msg, error) {
	for _, v := range app.handlers {
		if v.Match(in) {
			return v.Exchange(in)
		}
	}
	return nil, errors.New("no valid handler")
}

// Handler is ...
type Handler struct {
	// Upstream is ...
	Upstream
	// Matchers is ...
	Matchers []Matcher
}

// Match is ...
func (h *Handler) Match(msg *dns.Msg) bool {
	for _, v := range h.Matchers {
		if v.Match(msg) {
			return true
		}
	}
	return false
}

var (
	_ caddy.App          = (*App)(nil)
	_ caddy.CleanerUpper = (*App)(nil)
	_ caddy.Provisioner  = (*App)(nil)
	_ caddy.Validator    = (*App)(nil)
)
