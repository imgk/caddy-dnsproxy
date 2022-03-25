package handler

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"unsafe"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/miekg/dns"

	"github.com/imgk/caddy-dnsproxy/app"
)

func init() {
	caddy.RegisterModule(Handler{})
}

// DefaultPrefix is ...
const DefaultPrefix = "/dns-query"

// Handler is ...
type Handler struct {
	// Prefix is ...
	Prefix string `json:"prefix,omitempty"`
	// BufferPool is ...
	*app.BufferPool

	up app.Upstream
}

// CaddyModule is ...
func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.dns_over_https",
		New: func() caddy.Module { return new(Handler) },
	}
}

// Provision is ...
func (m *Handler) Provision(ctx caddy.Context) error {
	if m.Prefix == "" {
		m.Prefix = DefaultPrefix
	}
	if !ctx.AppIsConfigured(app.CaddyAppID) {
		return errors.New("dnsproxy is not configured")
	}
	mod, err := ctx.App(app.CaddyAppID)
	if err != nil {
		return err
	}
	m.up = mod.(app.Upstream)
	m.BufferPool = mod.(*app.App).BufferPool()
	return nil
}

// ServeHTTP is ...
func (m *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if strings.HasPrefix(r.URL.Path, m.Prefix) {
		switch r.Method {
		case http.MethodGet:
			return m.serveGet(w, r)
		case http.MethodPost:
			return m.servePost(w, r)
		default:
			return next.ServeHTTP(w, r)
		}
	}
	return next.ServeHTTP(w, r)
}

func (m *Handler) serveGet(w http.ResponseWriter, r *http.Request) error {
	ss, ok := r.URL.Query()["dns"]
	if !ok || len(ss) < 1 {
		return errors.New("no dns query")
	}

	buf := m.Get()
	defer m.Put(buf)

	n, err := base64.RawURLEncoding.Decode(buf, func(s string) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&s)))), len(s))
	}(ss[0]))
	if err != nil {
		return err
	}

	return m.response(w, buf, n)
}

func (m *Handler) servePost(w http.ResponseWriter, r *http.Request) error {
	buf := m.Get()
	defer m.Put(buf)

	// read dns message from request
	n, err := Buffer(buf).ReadFrom(r.Body)
	if err != nil {
		return err
	}

	return m.response(w, buf, int(n))
}

func (m *Handler) response(w http.ResponseWriter, buf []byte, n int) error {
	// parse dns message
	msg := &dns.Msg{}
	if err := msg.Unpack(buf[:n]); err != nil {
		return err
	}

	// request response
	msg, err := m.up.Exchange(msg)
	if err != nil {
		return err
	}

	bb, err := msg.PackBuffer(buf)
	if err != nil {
		return err
	}

	// write response back
	_, err = w.Write(bb)
	return err
}

var _ caddyhttp.MiddlewareHandler = (*Handler)(nil)

// Buffer is ...
type Buffer []byte

// ReadFrom is ...
func (b Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		nr, er := r.Read(b[n:])
		if nr > 0 {
			n += int64(nr)
		}
		if er != nil {
			if errors.Is(er, io.EOF) {
				break
			}
			err = er
			break
		}
		if int(n) == len(b) {
			err = io.ErrShortBuffer
			break
		}
	}
	return
}
