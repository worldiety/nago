package tdb

import (
	"bytes"
	"context"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
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
		for entry := range b.db.Range(b.bucket, minK, maxK) {
			if !yield(entry.key, nil) {
				return
			}
		}
	}

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
