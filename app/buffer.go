package app

import (
	"sync"

	"github.com/miekg/dns"

	"github.com/imgk/caddy-dnsproxy/buffer"
)

// BufferPool is ...
type BufferPool struct {
	Pool buffer.BufferPool[[]byte]
}

// NewBufferPool is ...
func NewBufferPool() *BufferPool {
	return &BufferPool{
		Pool: buffer.BufferPool[[]byte]{
			Pool: sync.Pool{
				New: NewBuffer,
			},
		},
	}
}

// Get is ...
func (p *BufferPool) Get() []byte {
	return *(p.Pool.Get())
}

// Put is ...
func (p *BufferPool) Put(b []byte) {
	p.Pool.Put(&b)
}

// NewBuffer is ...
func NewBuffer() any {
	buf := make([]byte, dns.MaxMsgSize)
	return &buf
}
