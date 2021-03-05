package tool

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"sync"
	"time"
)

type BufferedStreamReader struct {
	rd        io.Reader // reader provided by the client
	buffer    []byte
	readIndex int
	m         sync.Mutex
	err       error
	isClosed  bool
}

const chunkSize = 8192

// NewBufferedStreamReader returns a new Reader
func NewBufferedStreamReader(rd io.Reader, size int, firstChunkSize int) *BufferedStreamReader {
	r := &BufferedStreamReader{
		rd:     rd,
		buffer: make([]byte, 0, size),
	}
	go func() {
		var maxIndex int

		for {
			r.m.Lock()

			if r.isClosed {
				r.err = r.Close()
				r.m.Unlock()
				return
			}

			if len(r.buffer) == 0 {
				maxIndex = firstChunkSize
			} else {
				maxIndex = len(r.buffer) + chunkSize
			}
			if maxIndex > cap(r.buffer) {
				maxIndex = cap(r.buffer)
			}

			startIndex := len(r.buffer)

			r.m.Unlock()
			n, err := rd.Read(r.buffer[startIndex:maxIndex])
			r.m.Lock()

			r.buffer = r.buffer[:startIndex+n]
			if err != nil {
				r.err = err
				r.m.Unlock()
				return
			}

			r.m.Unlock()
		}
	}()

	return r
}

// Read reads data into p.
// It returns the number of bytes read into p.
// The bytes are taken from at most one Read on the underlying Reader,
// hence n may be less than len(p).
// At EOF, the count will be zero and err will be io.EOF.
func (b *BufferedStreamReader) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()

	if b.isClosed {
		return 0, os.ErrClosed
	}

	if b.rd == nil {
		return 0, errors.New("underlying reader is undefined")
	}

	if len(p) == 0 {
		return 0, nil
	}

	if b.readIndex >= cap(b.buffer) {
		return 0, io.EOF
	}

	for {
		max := b.readIndex + len(p)
		if max > cap(b.buffer) {
			max = cap(b.buffer)
		}

		if max <= len(b.buffer) {
			break
		}

		if b.err != nil {
			// No need to wait if there is an error
			return 0, b.err
		} else {
			b.m.Unlock()
			// Waiting stream sync
			time.Sleep(200 * time.Millisecond)
			b.m.Lock()
		}
	}

	n = copy(p, b.buffer[b.readIndex:])
	b.readIndex += n

	return n, err
}

func (b *BufferedStreamReader) Seek(offset int64, whence int) (int64, error) {
	b.m.Lock()
	defer b.m.Unlock()

	if b.isClosed {
		return 0, os.ErrClosed
	}

	var abs int64
	switch whence {
	case 0:
		abs = offset
	case 1:
		abs = int64(b.readIndex) + offset
	case 2:
		abs = int64(cap(b.buffer)) + offset
	default:
		return 0, errors.New("bufferedReader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("bufferedReader.Seek: negative position")
	}
	b.readIndex = int(abs)
	return abs, nil
}

func (b *BufferedStreamReader) Close() error {
	b.m.Lock()
	defer b.m.Unlock()

	b.isClosed = true
	return nil
}
