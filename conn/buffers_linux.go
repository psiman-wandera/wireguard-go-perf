// +build !android

package conn

import (
	"bytes"
	"errors"
)

type buffers struct {
	buffs    []*bytes.Buffer
	len, cap int
}

func buffersWithCap(cap int) buffers {
	b := buffers{
		cap: cap,
		len: 0,
	}
	for i := 0; i < cap; i++ {
		b.buffs = append(b.buffs, bytes.NewBuffer(make([]byte, 0, gsoSegmentSize)))
	}
	return b
}

func (b *buffers) Write(p []byte) (n int, err error) {
	if b.len < b.cap {
		b.buffs[b.len].Write(p)
		b.len++
		return len(p), nil
	}
	return 0, errors.New("buffers overflow")
}

func (b *buffers) Reset() {
	for _, b := range b.buffs {
		b.Reset()
	}
	b.len = 0
}

func (b *buffers) Bytes(i int) []byte { return b.buffs[i].Bytes() }

func (b *buffers) Len() int {
	return b.len
}

func (b *buffers) Cap() int {
	return b.cap
}
