package app

import (
	"crypto/tls"
	"errors"
	"io"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"

	"github.com/lucas-clemente/quic-go"
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
			App:  app,
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
			App:      app,
			lg:       app.Logger().Named("tcp"),
		}
		s.lg.Info("start server")
		return s, nil
	case "tls":
		connPolicies := caddytls.ConnectionPolicies{new(caddytls.ConnectionPolicy)}
		if err := connPolicies.Provision(ctx); err != nil {
			return nil, err
		}
		ln, err := caddy.Listen("tcp", ":"+strconv.Itoa(app.ListenTLS))
		if err != nil {
			return nil, err
		}
		ln = tls.NewListener(ln, connPolicies.TLSConfig(ctx))
		s := &Stream{
			Listener: ln,
			App:      app,
			lg:       app.Logger().Named("tls"),
		}
		s.lg.Info("start server")
		return s, nil
	case "quic":
		conn, err := caddy.ListenPacket("udp", ":"+strconv.Itoa(app.ListenQuic))
		if err != nil {
			return nil, err
		}
		connPolicies := caddytls.ConnectionPolicies{new(caddytls.ConnectionPolicy)}
		if err := connPolicies.Provision(ctx); err != nil {
			return nil, err
		}
		ln, err := quic.Listen(conn, connPolicies.TLSConfig(ctx), &quic.Config{})
		s := &Quic{
			Listener: ln,
			App:      app,
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
