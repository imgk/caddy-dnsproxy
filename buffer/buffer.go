package buffer

import "sync"

// BufferPool is ...
type BufferPool[T any] struct {
	sync.Pool
}

// Get is ...
func (p *BufferPool[T]) Get() *T {
	return p.Pool.Get().(*T)
}

// Put is ...
func (p *BufferPool[T]) Put(t *T) {
	p.Pool.Put(t)
}
