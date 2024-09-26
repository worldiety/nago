package pebble

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
	"time"
)

type BlobStore struct {
	db     *pebble.DB
	ticker *time.Ticker
	done   chan bool
	closed bool
	prefix []byte
}

func NewBlobStore(db *pebble.DB) *BlobStore {
	return &BlobStore{db: db}
}

func Open(dir string) (*BlobStore, error) {
	db, err := pebble.Open(dir, &pebble.Options{})
	if err != nil {
		return nil, err
	}

	return NewBlobStore(db), nil
}

// SetPrefix sets a store global prefix, e.g. if you want to partition the store in a transparent way, the given
// prefix is added and removed automatically.
func (b *BlobStore) SetPrefix(prefix string) {
	b.prefix = []byte(prefix)
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	prefix := append([]byte(opts.Prefix), b.prefix...)
	lower := append(prefix, opts.MinInc...)
	upper := append(prefix, opts.MaxInc...)
	if opts.MaxInc == "" {
		upper = append(upper, 0xFF)
	}

	it, err := b.db.NewIterWithContext(ctx, &pebble.IterOptions{
		LowerBound: lower,
		UpperBound: upper,
	})

	if err != nil {
		return func(yield func(string, error) bool) {
			yield("", fmt.Errorf("cannot create iter: %w", err))
		}
	}

	return func(yield func(string, error) bool) {
		isBreak := false
		for it.First(); it.Valid(); it.Next() {
			key := it.Key()
			if !bytes.HasPrefix(key, prefix) {
				break
			}

			if bytes.Compare(key, upper) >= 0 {
				break
			}

			if !yield(string(key[len(b.prefix):]), nil) {
				isBreak = true
				break
			}
		}

		err := it.Close()
		if !isBreak && err != nil {
			yield("", err)
		}
	}

}

func (b *BlobStore) keyWithPrefix(key string) []byte {
	if len(b.prefix) == 0 {
		return []byte(key)
	}

	return append(b.prefix, []byte(key)...)
}

func (b *BlobStore) Exists(ctx context.Context, key string) (bool, error) {
	_, closer, err := b.db.Get(b.keyWithPrefix(key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	defer closer.Close()

	return true, nil
}

func (b *BlobStore) Delete(ctx context.Context, key string) error {
	return b.db.Delete(b.keyWithPrefix(key), pebble.NoSync)
}

func (b *BlobStore) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	var res std.Option[io.ReadCloser]

	leakedBuf, closer, err := b.db.Get(b.keyWithPrefix(key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return res, nil
		}

		return res, err
	}

	// start the transaction and keep it alive until the reader is closed.
	// this will avoid another full allocation of a byte slice which is mostly very short-lived just
	// for unmarshalling json data.

	return std.Some[io.ReadCloser](&readerCloser{
		Reader: bytes.NewReader(leakedBuf),
		closer: closer,
	}), nil
}

func (b *BlobStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return &writeCloser{
		parent: b,
		Buffer: &bytes.Buffer{},
		key:    key,
		ctx:    ctx,
	}, nil
}

// Close only performs an actual close, if the Store was created using [Open], otherwise it is just a no-op.
func (b *BlobStore) Close() error {
	if b.closed {
		return nil
	}

	// check if we own the store
	if b.ticker != nil {
		b.ticker.Stop()
		b.done <- true

		// not sure if we need to sync, the code seems to be different
		if err := b.db.Close(); err != nil {
			return err
		}
	}

	b.closed = true

	return nil
}
