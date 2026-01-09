// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mem

import (
	"bytes"
	"context"
	"io"
	"iter"
	"slices"
	"sort"
	"strings"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xmaps"
)

var _ blob.Store = (*BlobStore)(nil)

// BlobStore provides an in-memory implementation without transactions.
// The transactions are just fake implementations to satisfy the contract and respect the read/write property.
// However, the store itself is at least thread safe.
type BlobStore struct {
	name   string
	values *xmaps.ConcurrentMap[string, []byte]
}

func (b *BlobStore) Name() string {
	return b.name
}

// NewBlobStore creates a new in-memory store.
func NewBlobStore(name string) *BlobStore {
	return &BlobStore{name: name, values: xmaps.NewConcurrentMap[string, []byte]()}
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var keys []string
		for k, _ := range b.values.All() {
			if opts.Prefix != "" {
				if !strings.HasPrefix(k, opts.Prefix) {
					continue
				}
			}

			if opts.MinInc != "" {
				if k < opts.MinInc {
					continue
				}
			}

			if opts.MaxInc != "" {
				if k > opts.MaxInc {
					continue
				}
			}

			if opts.Limit > 0 && len(keys) == opts.Limit {
				break
			}

			keys = append(keys, k)
		}

		sort.Strings(keys)

		if opts.Reverse {
			slices.Reverse(keys)
		}
		
		for _, key := range keys {
			if !yield(key, nil) {
				return
			}
		}

	}
}

func (b *BlobStore) Has(key string) bool {
	_, ok := b.values.Load(key)
	return ok
}

func (b *BlobStore) Load(key string) ([]byte, bool) {
	buf, ok := b.values.Load(key)
	return slices.Clone(buf), ok
}

func (b *BlobStore) Store(key string, buf []byte) {
	b.values.Store(key, slices.Clone(buf)) // defensive copy
}

func (b *BlobStore) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := b.values.Load(key)
	return ok, nil
}

func (b *BlobStore) Delete(ctx context.Context, key string) error {
	b.values.Delete(key)
	return nil
}

func (b *BlobStore) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	buf, ok := b.values.Load(key)
	if !ok {
		return std.None[io.ReadCloser](), nil
	}

	return std.Some[io.ReadCloser](readerCloser{bytes.NewReader(buf)}), nil
}

func (b *BlobStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return &writer{
		parent: b,
		key:    key,
	}, nil
}
