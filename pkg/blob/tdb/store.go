// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bytes"
	"context"
	"io"
	"iter"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
)

type BlobStore struct {
	db     *DB
	bucket string
}

func NewBlobStore(db *DB, bucket string) *BlobStore {
	return &BlobStore{
		db:     db,
		bucket: bucket,
	}
}

func (b *BlobStore) Name() string {
	return b.bucket
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		minK := opts.Prefix + opts.MinInc
		maxK := opts.Prefix + opts.MaxInc
		if opts.Prefix != "" && len(opts.MaxInc) == 0 {
			maxK = nextPrefix(maxK)
		}

		for entry := range b.db.Range(b.bucket, minK, maxK) {
			if !yield(entry.key, nil) {
				return
			}
		}
	}

}

func nextPrefix(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] < 0xFF {
			return s[:i] + string(s[i]+1)
		}
	}

	return ""
}

func (b *BlobStore) Count(ctx context.Context) (int64, error) {
	return int64(b.db.Len(b.bucket)), nil
}

func (b *BlobStore) Exists(ctx context.Context, key string) (bool, error) {
	return b.db.Exists(b.bucket, key), nil
}

func (b *BlobStore) Delete(ctx context.Context, key string) error {
	return b.db.Delete(b.bucket, key)
}

func (b *BlobStore) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	return b.db.Get(b.bucket, key), nil
}

func (b *BlobStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return &writeCloser{
		parent: b,
		Buffer: &bytes.Buffer{},
		key:    key,
		ctx:    ctx,
	}, nil
}

func (b *BlobStore) Close() error {
	return nil
}

type writeCloser struct {
	parent *BlobStore
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

	err := w.parent.db.Set(w.parent.bucket, w.key, w.Bytes())

	if err != nil {
		return err
	}

	w.closed = true

	return nil
}
