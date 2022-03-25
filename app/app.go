package app

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/caddyserver/caddy/v2"

	"github.com/miekg/dns"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// CaddyAppID is ...
const CaddyAppID = "dnsproxy"

const (
	// DefaultUDPPort is ...
	DefaultUDPPort = 53
	// DefaultTCPPort is ...
	DefaultTCPPort = 53
	// DefaultTLSPort is ...
	DefaultTLSPort = 853
	// DefaultQuicPort is ...
	DefaultQuicPort = 853
)

func init() {
	caddy.RegisterModule(App{})
}

// BufferPool is ...
type BufferPool struct {
	sync.Pool
}

// NewBufferPool is ...
func NewBufferPool() *BufferPool {
	return &BufferPool{
		Pool: sync.Pool{
			New: NewBuffer,
		},
	}
}

// NewBuffer is ...
func NewBuffer() any {
	buf := make([]byte, dns.MaxMsgSize)
	return &buf
}

// Get is ...
func (p *BufferPool) Get() []byte {
	return *(p.Pool.Get().(*[]byte))
}

// Put is ...
func (p *BufferPool) Put(buf []byte) {
	p.Pool.Put(&buf)
}

// App is ...
type App struct {
	// Handlers is ...
	Handlers []*struct {
		// UpstreamRaw is ...
		UpstreamRaw json.RawMessage `json:"upstream" caddy:"namespace=dnsproxy.upstreams inline_key=upstream"`
		// MatchersRaw is ...
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
	servers  []Server
	bp       *BufferPool
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
	app.bp = NewBufferPool()

	for _, v := range app.Servers {
		switch v {
		case "tcp", "udp", "tls", "quic":
			srv, err := NewServer(app, ctx, v)
			if err != nil {
				return err
			}
			app.servers = append(app.servers, srv)
		default:
			return errors.New("not a valid server type")
		}
	}

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
	for _, srv := range app.servers {
		go srv.Run()
	}
	return nil
}

// Stop is ...
func (app *App) Stop() error {
	errs := []error{}
	for _, srv := range app.servers {
		if err := srv.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

// Cleanup is ...
func (app *App) Cleanup() error {
	errs := []error{}
	for _, v := range app.handlers {
		if err := v.Cleanup(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

// Logger is ...
func (app *App) Logger() *zap.Logger {
	return app.lg
}

// BufferPool is ...
func (app *App) BufferPool() *BufferPool {
	return app.bp
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

// Cleanup is ...
func (h *Handler) Cleanup() error {
	errs := []error{}
	for _, vv := range h.Matchers {
		if cu, ok := vv.(caddy.CleanerUpper); ok {
			if err := cu.Cleanup(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) != 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

var (
	_ caddy.App          = (*App)(nil)
	_ caddy.CleanerUpper = (*App)(nil)
	_ caddy.Provisioner  = (*App)(nil)
	_ caddy.Validator    = (*App)(nil)
)
