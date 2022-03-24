package app

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// Stream is ...
type Stream struct {
	// Listener is ...
	net.Listener
	// App is ...
	App *App

	lg *zap.Logger
}

// Run is ..
func (s *Stream) Run() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.lg.Error(fmt.Sprintf("tcp/tls server error: accept error: %v", err))
			return
		}
		go func(conn net.Conn, app *App) {
			defer conn.Close()

			buf := make([]byte, 2048)
			msg := &dns.Msg{}
			for {
				if err := conn.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: net.Conn.SetReadDeadline error: %v", err))
					return
				}
				_, err := io.ReadFull(conn, buf[:2])
				if err != nil {
					if ne, ok := err.(net.Error); errors.Is(err, io.EOF) || (ok && ne.Timeout()) {
						return
					}
					s.lg.Error(fmt.Sprintf("tcp/tls server error: read length error: %v", err))
					return
				}
				n := int(buf[0])<<8 | int(buf[1])
				if _, err := io.ReadFull(conn, buf[:n]); err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: read message error: %v", err))
					return
				}
				if err := msg.Unpack(buf[:n]); err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: unpack error: %v", err))
					return
				}
				msg, err = app.Exchange(msg)
				if err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: exchange error: %v", err))
					return
				}
				bb, err := msg.PackBuffer(buf)
				if err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: pack error: %v", err))
					return
				}
				if _, err := conn.Write([]byte{byte(len(bb) >> 8), byte(len(bb))}); err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: write length error: %v", err))
					return
				}
				if _, err := conn.Write(bb); err != nil {
					s.lg.Error(fmt.Sprintf("tcp/tls server error: write message error: %v", err))
					return
				}
			}
		}(conn, s.App)
	}
	return
}

// Close is ...
func (s *Stream) Close() error {
	return s.Listener.Close()
}
