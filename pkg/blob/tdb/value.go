package tdb

import (
	"fmt"
	"github.com/rogpeppe/go-internal/lockedfile"
	"io"
	"os"
)

type Value struct {
	f      *lockedfile.File
	offset int64
	len    uint32
}

func (v Value) Len() int {
	return int(v.len)
}

// NewReader returns a reader which gives access to the underlying value. This is valid, until the according WAL
// respective file is closed. Omitting a Close currently just keeps a pointer to the file, which may be closed
// but otherwise does not cause any serious leak. Opening a zero value returns an already closed reader.
func (v Value) NewReader() io.ReadCloser {
	return &valReader{ptr: v, closed: v.f == nil}
}

func (v Value) Copy(dst []byte) error {
	if len(dst) != int(v.len) {
		return fmt.Errorf("buffer size mismatch")
	}

	_, err := v.f.ReadAt(dst, int64(v.offset))
	return err
}

type valReader struct {
	ptr    Value
	offset int64
	closed bool
}

func (v *valReader) Read(p []byte) (n int, err error) {
	if v.closed {
		return 0, os.ErrClosed
	}

	if len(p) == 0 {
		return 0, nil
	}
	if v.offset == int64(v.ptr.len) {
		return 0, io.EOF
	}

	if len(p)+int(v.offset) < v.ptr.Len() {
		n, err := v.ptr.f.ReadAt(p, v.offset+v.ptr.offset)
		v.offset += int64(n)
		return n, err
	} else {
		// virtual EOF of value length
		p = p[:min(len(p), v.ptr.Len()-int(v.offset))]
		n, err := v.ptr.f.ReadAt(p, v.offset+v.ptr.offset)
		v.offset += int64(n)
		if err == nil {
			return n, io.EOF
		}

		return n, err
	}
}

func (v *valReader) Close() error {
	v.closed = true
	v.ptr = Value{}
	return nil
}
