package bolt

import (
	"bytes"
	"errors"
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"io/fs"
)

type BlobStore struct {
	db         *bbolt.DB
	bucketName []byte
}

func NewBlobStore(db *bbolt.DB, bucketName string) *BlobStore {
	return &BlobStore{db: db, bucketName: []byte(bucketName)}
}

func (b *BlobStore) Update(f func(blob.Tx) error) error {
	txn := &boltTx{parent: b}
	defer func() {
		txn.valid = false
	}()

	return b.db.Update(func(tx *bbolt.Tx) error {
		txn.tx = tx
		txn.valid = true
		return f(txn)
	})
}

func (b *BlobStore) View(f func(blob.Tx) error) error {
	txn := &boltTx{parent: b}
	defer func() {
		txn.valid = false
	}()

	return b.db.View(func(tx *bbolt.Tx) error {
		txn.tx = tx
		txn.valid = true
		return f(txn)
	})
}

type boltTx struct {
	parent *BlobStore
	tx     *bbolt.Tx
	valid  bool
}

func (b *boltTx) Each(yield func(blob.Entry, error) bool) {
	bucket := b.tx.Bucket(b.parent.bucketName)
	if bucket == nil {
		return
	}

	yieldStopped := false
	err := bucket.ForEach(func(k, v []byte) error {
		valid := true
		yieldStopped = !yield(blob.Entry{
			Key: string(k), // we cannot use unsafe, because consumer expects to let the string key escape safely
			Open: func() (io.ReadCloser, error) {
				if !valid {
					return nil, fmt.Errorf("open outside of according yield call is invalid")
				}

				return readerCloser{Reader: bytes.NewReader(v)}, nil
			},
		}, nil)
		valid = false

		if yieldStopped {
			return fs.SkipAll
		}

		return nil
	})

	// do not yield if iteration has been cancelled
	if err != nil && !errors.Is(err, fs.SkipAll) {
		yield(blob.Entry{}, err)
	}
}

func (b *boltTx) Delete(key string) error {
	bucket := b.tx.Bucket(b.parent.bucketName)
	if bucket == nil {
		return nil
	}

	return bucket.Delete([]byte(key))
}

func (b *boltTx) Put(entry blob.Entry) error {
	bucket := b.tx.Bucket(b.parent.bucketName)
	if bucket == nil {
		bck, err := b.tx.CreateBucketIfNotExists(b.parent.bucketName)
		if err != nil {
			return err
		}

		bucket = bck
	}

	reader, err := entry.Open()
	if err != nil {
		return err
	}

	buf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return bucket.Put([]byte(entry.Key), buf)
}

func (b *boltTx) Get(key string) (std.Option[blob.Entry], error) {
	bucket := b.tx.Bucket(b.parent.bucketName)
	if bucket == nil {
		return std.None[blob.Entry](), nil
	}

	buf := bucket.Get([]byte(key))
	if buf == nil {
		return std.None[blob.Entry](), nil
	}

	return std.Some(blob.Entry{
		Key: key,
		Open: func() (io.ReadCloser, error) {
			if !b.valid {
				return nil, fmt.Errorf("open call outside of transaction")
			}

			return readerCloser{Reader: bytes.NewReader(buf)}, nil
		},
	}), nil
}
