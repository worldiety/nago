package bolt

import (
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/persistence/kv"
)

type Store struct {
	db *bbolt.DB
}

func NewStore(db *bbolt.DB) *Store {
	return &Store{db: db}
}

func (b *Store) Update(f func(kv.Tx) error) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return f(&Tx{tx: tx})
	})
}

func (b *Store) View(f func(kv.Tx) error) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		return f(&Tx{tx: tx})
	})
}

type Tx struct {
	tx *bbolt.Tx
}

func (t *Tx) Each(f func(name []byte, c kv.Bucket) error) error {
	return t.tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
		return f(name, &Bucket{bucket: b})
	})
}

func (t *Tx) Bucket(name []byte) (kv.Bucket, error) {
	var b *bbolt.Bucket
	var err error
	if t.Writable() {
		b, err = t.tx.CreateBucketIfNotExists(name)
	} else {
		b = t.tx.Bucket(name)
	}

	if err != nil {
		return nil, err
	}

	if b == nil {
		return nil, nil
	}

	return &Bucket{bucket: b}, nil
}

func (t *Tx) DeleteBucket(name []byte) error {
	return t.tx.DeleteBucket(name)
}

func (t *Tx) Writable() bool {
	return t.tx.Writable()
}

type Bucket struct {
	bucket *bbolt.Bucket
}

func (b *Bucket) Each(f func(key []byte, value []byte) error) error {
	return b.bucket.ForEach(f)
}

func (b *Bucket) Delete(key []byte) error {
	return b.bucket.Delete(key)
}

func (b *Bucket) Put(key, value []byte) error {
	return b.bucket.Put(key, value)
}

func (b *Bucket) Get(key []byte) ([]byte, error) {
	return b.bucket.Get(key), nil
}
