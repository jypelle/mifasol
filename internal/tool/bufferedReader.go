package tool

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"sync"
)

type BufferedReader struct {
	rd     io.Reader // reader provided by the client
	buffer *bytes.Buffer
	m      sync.Mutex
}

// NewBufferedReader returns a new Reader whose buffer has at least the specified
// size
func NewBufferedReader(size int) *BufferedReader {
	r := &BufferedReader{
		buffer: bytes.NewBuffer(make([]byte, size)),
	}
	return r
}

// Reset discards any buffered data, resets all state, and switches
// the buffered reader to read from r.
func (b *BufferedReader) Reset(r io.Reader) {
	b.m.Lock()
	defer b.m.Unlock()
	b.rd = r // reader provided by the client
	b.buffer.Reset()
	io.Copy(b.buffer, b.rd)
}

// Read reads data into p.
// It returns the number of bytes read into p.
// The bytes are taken from at most one Read on the underlying Reader,
// hence n may be less than len(p).
// To read exactly len(p) bytes, use io.ReadFull(b, p).
// At EOF, the count will be zero and err will be io.EOF.
func (b *BufferedReader) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	if b.rd == nil {
		return 0, errors.New("underlying reader is undefined")
	}
	return b.buffer.Read(p)
}

func (b *BufferedReader) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (b *BufferedReader) Close() error { return nil }
