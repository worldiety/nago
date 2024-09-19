package badger

import (
	"bytes"
	"context"
	"github.com/dgraph-io/badger/v4"
)

type writeCloser struct {
	db *badger.DB
	*bytes.Buffer
	closed bool
	key    string // conversion inline below is probably GC free, inlined and optimized away
	ctx    context.Context
}

func (w *writeCloser) Close() error {
	if w.closed {
		return nil
	}

	// check if the context was cancelled, so that we don't commit unwanted stuff
	if w.ctx.Err() != nil {
		return w.ctx.Err()
	}

	// using this approach, we can guarantee deadlock-free behavior, because a write transaction does never need
	// to await for something external.
	err := w.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(w.key), w.Buffer.Bytes())
	})

	if err != nil {
		return err
	}

	w.closed = true

	return nil
}
