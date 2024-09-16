package mem

import "bytes"

type writer struct {
	parent *BlobStore
	key    string
	closed bool
	bytes.Buffer
}

func (w *writer) Close() error {
	if w.closed {
		return nil
	}

	w.closed = true
	w.parent.values.Store(w.key, w.Buffer.Bytes())
	return nil
}
