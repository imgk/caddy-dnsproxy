package buffer

import "testing"

func TestBufferPool(t *testing.T) {
	p := NewBufferPool[[]byte](func() any {
		buf := make([]byte, 512, 1024)
		return &buf
	})

	ptr := p.Get()
	defer p.Put(ptr)

	if len(*ptr) != 512 || cap(*ptr) != 1024 {
		t.Error("test buffer error")
	}
}
