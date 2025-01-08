package crypto

import (
	"bytes"
	"context"
	"fmt"
)

type writer struct {
	parent storeAdapter
	key    string
	closed bool
	bytes.Buffer
	ctx context.Context
}

func (w *writer) Close() error {
	if w.closed {
		return nil
	}

	w.closed = true
	cypher, err := w.parent.encode(w.Bytes())
	if err != nil {
		return fmt.Errorf("cannot encrypt buffer: %w", err)
	}

	wrt, err := w.parent.delegate.NewWriter(w.ctx, w.key)
	if err != nil {
		return fmt.Errorf("cannot delegate writer: %w", err)
	}

	if _, err := wrt.Write(cypher); err != nil {
		return fmt.Errorf("cannot write buffer: %w", err)
	}

	if err := wrt.Close(); err != nil {
		return fmt.Errorf("cannot commit writer: %w", err)
	}
	
	return nil
}
