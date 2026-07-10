// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package checkpoint

import (
	"encoding/binary"
	"fmt"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/ndb"
)

// blobStore is the default [Store]: it persists the cursor as an 8-byte
// big-endian [ndb.Seq] under a single blob key.
type blobStore struct {
	store blob.Store
	key   string
}

// NewBlobStore returns a [Store] backed by a single blob key in store. The
// cursor is stored as an 8-byte big-endian [ndb.Seq]. A missing or empty blob
// reads back as 0 ("from the beginning").
func NewBlobStore(store blob.Store, key string) Store {
	return blobStore{store: store, key: key}
}

func (b blobStore) Load() (ndb.Seq, error) {
	opt, err := blob.Get(b.store, b.key)
	if err != nil {
		return 0, fmt.Errorf("checkpoint: load cursor %q: %w", b.key, err)
	}

	if opt.IsNone() {
		return 0, nil
	}

	buf := opt.Unwrap()
	if len(buf) == 0 {
		return 0, nil
	}
	if len(buf) != 8 {
		return 0, fmt.Errorf("checkpoint: cursor %q has unexpected length %d (want 8)", b.key, len(buf))
	}

	return ndb.Seq(binary.BigEndian.Uint64(buf)), nil
}

func (b blobStore) Save(seq ndb.Seq) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(seq))
	if err := blob.Put(b.store, b.key, buf[:]); err != nil {
		return fmt.Errorf("checkpoint: save cursor %q: %w", b.key, err)
	}
	return nil
}
