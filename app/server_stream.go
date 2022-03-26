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
	// BufferPool is ...
	*BufferPool

	up Upstream
	lg *zap.Logger
}

// Run is ..
func (s *Stream) Run() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.lg.Error(fmt.Sprintf("server error: accept error: %v", err))
			return
		}
		// handle net.Conn
		go func(conn net.Conn, up Upstream) {
			defer conn.Close()

			ptr, buf := s.GetValue()
			defer s.Put(ptr)

			msg := &dns.Msg{}

			// handle message loop
			for {
				if err := conn.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
					s.lg.Error(fmt.Sprintf("server error: net.Conn.SetReadDeadline error: %v", err))
					return
				}

				// read prefix
				if _, err := io.ReadFull(conn, buf[:2]); err != nil {
					if ne, ok := err.(net.Error); errors.Is(err, io.EOF) || (ok && ne.Timeout()) {
						return
					}
					s.lg.Error(fmt.Sprintf("server error: read length error: %v", err))
					return
				}

				// read message
				n := int(buf[0])<<8 | int(buf[1])
				if _, err := io.ReadFull(conn, buf[:n]); err != nil {
					s.lg.Error(fmt.Sprintf("server error: read message error: %v", err))
					return
				}
				if err := msg.Unpack(buf[:n]); err != nil {
					s.lg.Error(fmt.Sprintf("server error: unpack error: %v", err))
					return
				}

				// request response
				msg, err = up.Exchange(msg)
				if err != nil {
					s.lg.Error(fmt.Sprintf("server error: exchange error: %v", err))
					return
				}
				bb, err := msg.PackBuffer(buf)
				if err != nil {
					s.lg.Error(fmt.Sprintf("server error: pack error: %v", err))
					return
				}

				// write prefix and message
				if _, err := conn.Write([]byte{byte(len(bb) >> 8), byte(len(bb))}); err != nil {
					s.lg.Error(fmt.Sprintf("server error: write length error: %v", err))
					return
				}
				if _, err := conn.Write(bb); err != nil {
					s.lg.Error(fmt.Sprintf("server error: write message error: %v", err))
					return
				}
			}
		}(conn, s.up)
	}
	return
}

// Close is ...
func (s *Stream) Close() error {
	return s.Listener.Close()
}
