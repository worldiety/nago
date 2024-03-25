package mem

import "bytes"

type readerCloser struct {
	*bytes.Reader
}

func (readerCloser) Close() error {
	return nil
}
