// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

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
