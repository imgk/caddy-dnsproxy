package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/imgk/memory-go"
)

// NextProtoDQ is the ALPN token for DoQ. During connection establishment,
// DNS/QUIC support is indicated by selecting the ALPN token "dq" in the
// crypto handshake.
// Current draft version:
// https://datatracker.ietf.org/doc/html/draft-ietf-dprive-dnsoquic-02
const NextProtoDQ = "doq-i02"

// Quic is ...
type Quic struct {
	// Listener is ...
	Listener *quic.Listener

	up Upstream
	lg *zap.Logger
}

// Run is ...
func (s *Quic) Run() {
	// accept new session
	for {
		sess, err := s.Listener.Accept(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "server closed") {
				return
			}
			s.lg.Error(fmt.Sprintf("accept session error: %v", err))
			return
		}
		go s.handleSession(sess)
	}
}

// CLose is ...
func (s *Quic) Close() error {
	return s.Listener.Close()
}

func (s *Quic) handleSession(sess quic.Connection) {
	defer sess.CloseWithError(0, "")

	// accept new stream
	for {
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			s.lg.Error(fmt.Sprintf("accept stream error: %v", err))
			return
		}
		go s.handleStream(stream)
	}
}

func (s *Quic) handleStream(stream quic.Stream) {
	defer stream.Close()

	ptr, buf := memory.Alloc[byte](dns.MaxMsgSize)
	defer memory.Free(ptr)

	msg := &dns.Msg{}

	// read message
	n, err := stream.Read(buf)
	if err != nil {
		s.lg.Error(fmt.Sprintf("server error: read stream error: %v", err))
		return
	}
	if err := msg.Unpack(buf[:n]); err != nil {
		s.lg.Error(fmt.Sprintf("server error: unpack error: %v", err))
		return
	}

	// request response
	msg, err = s.up.Exchange(msg)
	if err != nil {
		s.lg.Error(fmt.Sprintf("server error: exchange error: %v", err))
		return
	}
	bb, err := msg.PackBuffer(buf)
	if err != nil {
		s.lg.Error(fmt.Sprintf("server error: pack error: %v", err))
		return
	}

	// write message
	if _, err := stream.Write(bb); err != nil {
		s.lg.Error(fmt.Sprintf("server error: write error: %v", err))
		return
	}
}
