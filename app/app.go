package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const CaddyAppID = "dnsproxy"

func init() {
	caddy.RegisterModule(App{})
}

// App is ...
type App struct {
	Handlers []*struct {
		Upstream    string            `json:"upstream"`
		Bootstrap   string            `json:"bootstrap,omitempty"`
		Timeout     caddy.Duration    `json:"timeout,omitempty"`
		MatchersRaw []json.RawMessage `json:"match" caddy:"namespace=dnsproxy.matchers inline_key=matcher"`
	} `json:"handlers"`

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
	app.lg = ctx.Logger(app)
	for _, v := range app.Handlers {
		hd := Handler{}

		// parse upstream
		up, err := upstream.AddressToUpstream(v.Upstream, nil)
		if err != nil {
			return err
		}
		hd.Upstream = up

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
	return nil, nil
}

// Handler is ...
type Handler struct {
	// Upstream is ...
	upstream.Upstream
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
