package badger

import (
	"bytes"
	"context"
	"errors"
	badger "github.com/dgraph-io/badger/v4"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
	"log/slog"
	"time"
)

var _ blob.Store = (*BlobStore)(nil)

type BlobStore struct {
	db     *badger.DB
	ticker *time.Ticker
	done   chan bool
	closed bool
}

// Open opens a directory with a badger database using a kind of eventual consistency.
// We are using asynchronous writes here, to maximize (fake) tps to the risk of loosing data.
// However, to lower the risk in the long term, we sync every 10 minutes, thus at most we loose the last
// 10 minutes of data. This is an intentional tradeoff regarding the costs of hosting. Most customers
// don't like to pay for this, because it may mean that even beefy machines with super fast SSDs can only
// reach 100-1000 TPS per machine, which actually hurts peak usage scenarios so badly, that regular
// fsync based systems become unusable at scale. This is also why we can't recommend bbolt, because
// without fsync it would corrupt itself by definition, which is by definition not the case for badger, which
// only looses latest data and not everything.
func Open(dir string) (*BlobStore, error) {
	opts := badger.DefaultOptions(dir)
	opts.SyncWrites = false // most important tradeoff, difference between 100 tps and ???
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	store := NewBlobStore(db)
	store.ticker = time.NewTicker(time.Minute * 10)
	store.done = make(chan bool)
	go func() {
		for {
			select {
			case <-store.done:
				return
			case <-store.ticker.C:
				if err := store.db.Sync(); err != nil {
					slog.Error("failed to sync badger db", slog.Any("err", err))
				}
			}
		}
	}()
	return store, nil
}

func NewBlobStore(db *badger.DB) *BlobStore {
	return &BlobStore{db: db}
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	var res []string
	err := b.db.View(func(txn *badger.Txn) error {
		bopts := badger.DefaultIteratorOptions
		bopts.PrefetchValues = false // optimize value reads away, we just need the key set here
		it := txn.NewIterator(bopts)
		defer it.Close()

		var zero blob.ListOptions
		if opts == zero {
			res = allKeys(it)
		} else {
			res = prefixAndRange(it, opts)
		}

		return nil
	})

	if err != nil {
		return func(yield func(string, error) bool) {
			yield("", err)
		}
	}

	return func(yield func(string, error) bool) {
		for _, key := range res {
			if !yield(key, nil) {
				return
			}
		}
	}
}

func prefixAndRange(it *badger.Iterator, opts blob.ListOptions) []string {
	var res []string

	var prefix []byte
	if len(opts.Prefix) > 0 {
		prefix = []byte(opts.Prefix)
	}
	var start []byte
	if len(opts.MinInc) > 0 {
		start = []byte(opts.MinInc)
	} else if len(opts.Prefix) > 0 {
		start = []byte(opts.Prefix)
	}

	var end []byte
	if len(opts.MaxInc) > 0 {
		end = []byte(opts.MaxInc)
	}

	for it.Seek(start); it.Valid(); it.Next() {
		item := it.Item()
		k := item.Key()
		if prefix != nil && !bytes.HasPrefix(k, prefix) {
			break
		}

		if end != nil && bytes.Compare(k, end) >= 0 {
			break
		}

		res = append(res, string(k))
	}

	return res
}

func allKeys(it *badger.Iterator) []string {
	var res []string

	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		k := item.Key()
		res = append(res, string(k))
	}

	return res
}

func (b *BlobStore) Exists(ctx context.Context, key string) (bool, error) {
	var found bool
	err := b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}

			return err
		}

		found = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return found, nil
}

func (b *BlobStore) Delete(ctx context.Context, key string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (b *BlobStore) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	var res std.Option[io.ReadCloser]

	// start the transaction and keep it alive until the reader is closed.
	// this will avoid another full allocation of a byte slice which is mostly very short-lived just
	// for unmarshalling json data.
	tx := b.db.NewTransaction(false)
	item, err := tx.Get([]byte(key))
	if err != nil {
		tx.Discard()

		if errors.Is(err, badger.ErrKeyNotFound) {
			return res, nil
		}

		return res, err
	}

	var leakedBuf []byte
	if err := item.Value(func(val []byte) error {
		leakedBuf = val
		return nil
	}); err != nil {
		return res, err
	}

	return std.Some[io.ReadCloser](&readerCloser{
		Reader: bytes.NewReader(leakedBuf),
		tx:     tx,
	}), nil
}

func (b *BlobStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return &writeCloser{
		db:     b.db,
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
