package app

import (
	"crypto/tls"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"

	"github.com/quic-go/quic-go"
)

// Server is ...
type Server interface {
	// Run is ...
	Run()
	// Close is ...
	io.Closer
}

// NewServer is ...
func NewServer(app *App, ctx caddy.Context, t string) (Server, error) {
	switch t {
	case "udp":
		conn, err := caddy.ListenPacket("udp", ":"+strconv.Itoa(app.ListenUDP))
		if err != nil {
			return nil, err
		}
		s := &Packet{
			Conn: conn,
			up:   app,
			lg:   app.Logger().Named("udp"),
		}
		s.lg.Info("start server")
		return s, nil
	case "tcp":
		ln, err := caddy.Listen("tcp", ":"+strconv.Itoa(app.ListenTCP))
		if err != nil {
			return nil, err
		}
		s := &Stream{
			Listener: ln,
			up:       app,
			lg:       app.Logger().Named("tcp"),
		}
		s.lg.Info("start server")
		return s, nil
	case "tls":
		// enable https
		connPolicies := caddytls.ConnectionPolicies{new(caddytls.ConnectionPolicy)}
		if err := connPolicies.Provision(ctx); err != nil {
			return nil, err
		}
		ln, err := caddy.Listen("tcp", ":"+strconv.Itoa(app.ListenTLS))
		if err != nil {
			return nil, err
		}
		// create tls.Config
		tlsConfig := connPolicies.TLSConfig(ctx)
		ln = tls.NewListener(ln, tlsConfig)
		s := &Stream{
			Listener: ln,
			up:       app,
			lg:       app.Logger().Named("tls"),
		}
		s.lg.Info("start server")
		return s, nil
	case "quic":
		conn, err := caddy.ListenPacket("udp", ":"+strconv.Itoa(app.ListenQuic))
		if err != nil {
			return nil, err
		}
		// enable https
		connPolicies := caddytls.ConnectionPolicies{new(caddytls.ConnectionPolicy)}
		if err := connPolicies.Provision(ctx); err != nil {
			return nil, err
		}
		// create tls.Config
		tlsConfig := connPolicies.TLSConfig(ctx)
		tlsConfig.NextProtos = []string{NextProtoDQ, "doq-i00", "dq", "doq"}
		ln, err := quic.Listen(conn, tlsConfig, &quic.Config{
			MaxIdleTimeout: 5 * time.Minute,
		})
		s := &Quic{
			Listener: ln,
			up:       app,
			lg:       app.Logger().Named("quic"),
		}
		s.lg.Info("start server")
		return s, nil
	default:
		return nil, errors.New("not a valid server type")
	}
}

var (
	_ Server = (*Packet)(nil)
	_ Server = (*Quic)(nil)
	_ Server = (*Stream)(nil)
)
