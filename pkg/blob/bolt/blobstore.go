package bolt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
)

var _ blob.Store = (*BlobStore)(nil)

type BlobStore struct {
	db         *bbolt.DB
	bucketName []byte
}

func NewBlobStore(db *bbolt.DB, bucketName string) *BlobStore {
	return &BlobStore{db: db, bucketName: []byte(bucketName)}
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	var res []string
	return func(yield func(string, error) bool) {
		err := b.db.View(func(tx *bbolt.Tx) error {
			if opts.Prefix != "" && (opts.MinInc != "" || opts.MaxInc != "") {
				return fmt.Errorf("applying prefix and range at the same time is not supported")
			}

			if opts.Prefix != "" {
				res = prefix(tx, b.bucketName, opts.Prefix, opts.Limit)
				return nil
			}

			if opts.MinInc != "" {
				res = ranger(tx, b.bucketName, opts.MinInc, opts.MaxInc, opts.Limit)
				return nil
			}

			// no filter case
			bucket := tx.Bucket(b.bucketName)
			if bucket == nil {
				return nil
			}

			err := bucket.ForEach(func(k, v []byte) error {
				res = append(res, string(k))
				if opts.Limit > 0 && len(res) >= opts.Limit {
					return data.SkipAll
				}

				return nil
			})

			if errors.Is(err, data.SkipAll) {
				return nil
			}

			return err
		})

		if err != nil {
			yield("", err)
			return
		}

		// we must not yield from the transaction, because the loop body of yield may start a write transaction
		// which cannot be upgraded and we get a deadlock per definition
		for _, key := range res {
			if !yield(key, nil) {
				return
			}
		}
	}
}

func (b *BlobStore) Exists(ctx context.Context, key string) (bool, error) {
	exists := false
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.bucketName)
		if bucket == nil {
			return nil
		}

		exists = bucket.Get([]byte(key)) != nil
		return nil
	})

	return exists, err
}

func (b *BlobStore) Delete(ctx context.Context, key string) error {
	err := b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.bucketName)
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})

	return err
}

// NewReader opens a read transaction and returns it through the given ReadCloser. You must
// ensure to close the reader, otherwise you will block any writes infinitely. Otherwise, it protects
// you against other issues arising with bbolt while mixing read and write transactions.
func (b *BlobStore) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	var res std.Option[io.ReadCloser]

	// start the transaction and keep it alive until the reader is closed.
	// this will avoid another full allocation of a byte slice which is mostly very short-lived just
	// for unmarshalling json data.
	tx, err := b.db.Begin(false)
	if err != nil {
		return res, err
	}

	bucket := tx.Bucket(b.bucketName)
	if bucket == nil {
		// even the bucket does not exist
		if err := tx.Rollback(); err != nil {
			panic(fmt.Errorf("unreachable: %w", err))
		}

		return res, nil
	}

	mmapBuf := bucket.Get([]byte(key))
	if mmapBuf == nil {
		// by definition of bbolt, this means that no such entry exists
		if err := tx.Rollback(); err != nil {
			panic(fmt.Errorf("unreachable: %w", err))
		}

		return res, nil
	}

	return std.Some[io.ReadCloser](&readerCloser{
		Reader: bytes.NewReader(mmapBuf),
		tx:     tx,
	}), nil
}

// NewWriter creates a write buffer which is written at one piece when closed, so you must check that error, to ensure
// persistence. This way, we can guarantee to be deadlock-free due to stalled bbolt transactions.
func (b *BlobStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	// this works differently than the read. First of all, we likely have a 10_000:1 read-write-ratio due to
	// the way the renderer works and secondly, delayed writes are deadlock prone and thirdly bbolt has no
	// streaming API anyway, thus we need a full buffered slice.

	return &writeCloser{
		bucketName: b.bucketName,
		db:         b.db,
		Buffer:     &bytes.Buffer{},
		key:        key,
		ctx:        ctx,
	}, nil
}

func ranger(tx *bbolt.Tx, bucketName []byte, lowInc, highInc string, limit int) []string {
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return nil
	}

	c := bucket.Cursor()
	min := []byte(lowInc)
	max := []byte(highInc)
	var res []string
	for k, _ := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, _ = c.Next() {
		res = append(res, string(k))
		if limit > 0 && len(res) >= limit {
			break
		}
	}

	return res
}

func prefix(tx *bbolt.Tx, bucketName []byte, prefix string, limit int) []string {
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return nil
	}

	c := bucket.Cursor()
	pfix := []byte(prefix)

	var res []string
	for k, _ := c.Seek(pfix); k != nil && bytes.HasPrefix(k, pfix); k, _ = c.Next() {
		res = append(res, string(k))
		if limit > 0 && len(res) >= limit {
			break
		}
	}

	return res
}
