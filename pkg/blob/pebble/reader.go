package pebble

import (
	"bytes"
	"io"
)

type readerCloser struct {
	*bytes.Reader
	closer io.Closer
	closed bool
}

func (r *readerCloser) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.Reader = nil

	return r.closer.Close()
}
