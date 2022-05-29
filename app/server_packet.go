package app

import (
	"errors"
	"fmt"
	"net"

	"github.com/imgk/memory-go"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// Packet is ...
type Packet struct {
	// Conn is ...
	Conn net.PacketConn

	up Upstream
	lg *zap.Logger
}

// Run is ..
func (s *Packet) Run() {
	ptr, buf := memory.Alloc[byte](dns.MaxMsgSize)
	defer memory.Free(ptr)

	msg := &dns.Msg{}

	for {
		// read message
		n, addr, err := s.Conn.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.lg.Error(fmt.Sprintf("server error: read packet error: %v", err))
			return
		}
		if err := msg.Unpack(buf[:n]); err != nil {
			s.lg.Error(fmt.Sprintf("server error: unpack error: %v", err))
			continue
		}

		// request response
		msg, err = s.up.Exchange(msg)
		if err != nil {
			s.lg.Error(fmt.Sprintf("server error: exchange error: %v", err))
			continue
		}
		bb, err := msg.PackBuffer(buf)
		if err != nil {
			s.lg.Error(fmt.Sprintf("server error: pack error: %v", err))
			continue
		}

		// write message
		if _, err := s.Conn.WriteTo(bb, addr); err != nil {
			s.lg.Error(fmt.Sprintf("server error: write back error: %v", err))
			continue
		}
	}
}

// Close is ...
func (s *Packet) Close() error {
	return s.Conn.Close()
}
