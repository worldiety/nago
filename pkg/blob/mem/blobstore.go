package mem

import (
	"bytes"
	"context"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xmaps"
	"io"
	"iter"
	"sort"
	"strings"
)

var _ blob.Store = (*BlobStore)(nil)

// BlobStore provides an in-memory implementation without transactions.
// The transactions are just fake implementations to satisfy the contract and respect the read/write property.
// However, the store itself is at least thread safe.
type BlobStore struct {
	values *xmaps.ConcurrentMap[string, []byte]
}

// NewBlobStore creates a new in-memory store.
func NewBlobStore() *BlobStore {
	return &BlobStore{values: xmaps.NewConcurrentMap[string, []byte]()}
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var keys []string
		for k, _ := range b.values.All {
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

			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, key := range keys {
			if !yield(key, nil) {
				return
			}
		}

	}
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
