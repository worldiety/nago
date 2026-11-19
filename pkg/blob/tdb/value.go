// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"fmt"
	"io"
	"os"
)

// Value refers to a stored value without holding it in memory.
type Value struct {
	f      *os.File
	offset int64
	len    uint32
}

func (v Value) Len() int {
	return int(v.len)
}

// NewReader returns a reader which gives access to the underlying value. It stays valid, even if a compaction
// replaces the file meanwhile, until the DB is closed. A reader keeps its file open, thus a reader of a replaced file
// should not be kept for a long time, but omitting a Close does not leak anything. Opening a zero value returns an
// already closed reader.
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
