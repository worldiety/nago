package badger

import (
	"bytes"
	"github.com/dgraph-io/badger/v4"
)

type readerCloser struct {
	*bytes.Reader
	tx     *badger.Txn
	closed bool
}

func (r *readerCloser) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.Reader = nil

	r.tx.Discard()

	return nil
}
