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
	"iter"
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

func (b *BlobStore) Get(key string) (std.Option[blob.Entry], error) {
	var res std.Option[blob.Entry]
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(b.bucketName)
		if bucket == nil {
			return nil
		}

		buf := bucket.Get([]byte(key))
		if buf == nil {
			return nil
		}

		tmp := bytes.Clone(buf) // buf is owned by tx
		res = std.Some[blob.Entry](blob.Entry{
			Key: key,
			Open: func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(tmp)), nil
			},
		})

		return nil
	})

	return res, err
}

func (b *BlobStore) Range(lowInc, highInc string) iter.Seq2[blob.Entry, error] {
	var collectedKeys []string // ensure that we never have a nested transaction, which may deadlock see https://github.com/etcd-io/bbolt?tab=readme-ov-file#transactions
	err := b.View(func(t blob.Tx) error {
		tx := t.(*boltTx)
		for entry, err := range tx.PrefixRange(lowInc, highInc) {
			if err != nil {
				panic(fmt.Errorf("unreachable: bbolt has no err in this code path, its mmapped"))
			}

			collectedKeys = append(collectedKeys, entry.Key)
		}

		return nil
	})

	return func(yield func(blob.Entry, error) bool) {
		if err != nil {
			yield(blob.Entry{}, nil)
			return
		}

		for _, key := range collectedKeys {
			entOpt, err := b.Get(key)
			if err != nil {
				if !yield(blob.Entry{}, err) {
					return
				}
			}

			if !yield(entOpt.Unwrap(), nil) {
				return
			}
		}
	}
}

func (b *BlobStore) Prefix(prefix string) iter.Seq2[blob.Entry, error] {
	//TODO implement me
	panic("implement me")
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

	defer reader.Close()

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

func (b *boltTx) PrefixRange(lowInc, highInc string) iter.Seq2[blob.Entry, error] {
	bucket := b.tx.Bucket(b.parent.bucketName)
	if bucket == nil {
		return func(yield func(blob.Entry, error) bool) {

		}
	}

	c := bucket.Cursor()
	min := []byte(lowInc)
	max := []byte(highInc)

	return func(yield func(blob.Entry, error) bool) {
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			yield(blob.Entry{
				Key: string(k),
				Open: func() (io.ReadCloser, error) {
					return readerCloser{Reader: bytes.NewReader(v)}, nil
				},
			}, nil)
		}
	}
}

func (b *boltTx) Prefix(prefix string) iter.Seq2[blob.Entry, error] {
	bucket := b.tx.Bucket(b.parent.bucketName)
	if bucket == nil {
		return func(yield func(blob.Entry, error) bool) {

		}
	}

	c := bucket.Cursor()
	pfix := []byte(prefix)

	return func(yield func(blob.Entry, error) bool) {
		for k, v := c.Seek(pfix); k != nil && bytes.HasPrefix(k, pfix); k, v = c.Next() {
			yield(blob.Entry{
				Key: string(k),
				Open: func() (io.ReadCloser, error) {
					return readerCloser{Reader: bytes.NewReader(v)}, nil
				},
			}, nil)
		}
	}
}
