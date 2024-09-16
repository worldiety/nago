package bolt

import (
	"bytes"
	"go.etcd.io/bbolt"
)

type readerCloser struct {
	*bytes.Reader
	tx     *bbolt.Tx
	closed bool
}

func (r *readerCloser) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.Reader = nil

	return r.tx.Rollback()
}
