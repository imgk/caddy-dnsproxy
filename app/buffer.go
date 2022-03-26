package app

import (
	"sync"

	"github.com/miekg/dns"

	"github.com/imgk/caddy-dnsproxy/buffer"
)

// BufferPool is ...
type BufferPool struct {
	buffer.BufferPool[[]byte]
}

// NewBufferPool is ...
func NewBufferPool() *BufferPool {
	return &BufferPool{
		BufferPool: buffer.BufferPool[[]byte]{
			Pool: sync.Pool{
				New: NewBuffer,
			},
		},
	}
}

// NewBuffer is ...
func NewBuffer() any {
	buf := make([]byte, dns.MaxMsgSize)
	return &buf
}
