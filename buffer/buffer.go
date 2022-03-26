package buffer

import (
	"errors"
	"log"
	"sync"
)

// BufferPool is ...
type BufferPool[T any] struct {
	sync.Pool
}

// NewBufferPool is ...
func NewBufferPool[T any](fn func() any) *BufferPool[T] {
	if _, ok := fn().(*T); !ok {
		log.Panic(errors.New("not a valid function for sync.Pool.New"))
	}
	return &BufferPool[T]{
		Pool: sync.Pool{
			New: fn,
		},
	}
}

// Get is ...
func (p *BufferPool[T]) Get() *T {
	return p.Pool.Get().(*T)
}

// GetValue is ...
func (p *BufferPool[T]) GetValue() (*T, T) {
	ptr := p.Pool.Get().(*T)
	return ptr, *ptr
}

// Put is ...
func (p *BufferPool[T]) Put(t *T) {
	p.Pool.Put(t)
}
