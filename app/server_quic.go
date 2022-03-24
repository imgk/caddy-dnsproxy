package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// Quic is ...
type Quic struct {
	// Listener is ...
	Listener quic.Listener
	// App is ...
	App *App

	lg *zap.Logger
}

// Run is ...
func (s *Quic) Run() {
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
	return
}

// CLose is ...
func (s *Quic) Close() error {
	return s.Listener.Close()
}

func (s *Quic) handleSession(sess quic.Session) {
	defer sess.CloseWithError(0, "")

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

	buf := make([]byte, 2048)
	msg := &dns.Msg{}

	n, err := stream.Read(buf)
	if err != nil {
		s.lg.Error(fmt.Sprintf("quic error: read stream error: %v", err))
		return
	}
	if err := msg.Unpack(buf[:n]); err != nil {
		s.lg.Error(fmt.Sprintf("quic error: quic error: unpack error: %v", err))
		return
	}
	msg, err = s.App.Exchange(msg)
	if err != nil {
		s.lg.Error(fmt.Sprintf("quic error: exchange error: %v", err))
		return
	}
	bb, err := msg.PackBuffer(buf)
	if err != nil {
		s.lg.Error(fmt.Sprintf("quic error: pack error: %v", err))
		return
	}
	if _, err := stream.Write(bb); err != nil {
		s.lg.Error(fmt.Sprintf("quic error: write error: %v", err))
		return
	}
}
