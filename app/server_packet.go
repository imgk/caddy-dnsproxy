package app

import (
	"errors"
	"fmt"
	"net"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// Packet is ...
type Packet struct {
	// Conn is ...
	Conn net.PacketConn
	// App is ...
	App *App

	lg *zap.Logger
}

// Run is ..
func (s *Packet) Run() {
	buf := make([]byte, 2048)
	msg := &dns.Msg{}
	for {
		n, addr, err := s.Conn.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.lg.Error(fmt.Sprintf("udp server error: read packet error: %v", err))
			return
		}
		if err := msg.Unpack(buf[:n]); err != nil {
			s.lg.Error(fmt.Sprintf("udp server error: unpack error: %v", err))
			continue
		}
		msg, err = s.App.Exchange(msg)
		if err != nil {
			s.lg.Error(fmt.Sprintf("udp server error: exchange error: %v", err))
			continue
		}
		bb, err := msg.PackBuffer(buf)
		if err != nil {
			s.lg.Error(fmt.Sprintf("udp server error: pack error: %v", err))
			continue
		}
		if _, err := s.Conn.WriteTo(bb, addr); err != nil {
			s.lg.Error(fmt.Sprintf("udp server error: write back error: %v", err))
			continue
		}
	}
	return
}

// Close is ...
func (s *Packet) Close() error {
	return s.Conn.Close()
}
