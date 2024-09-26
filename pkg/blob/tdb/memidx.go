package tdb

/*
import (
	"go.wdy.de/nago/pkg/xmaps"
	"os"
	"path/filepath"
)

type payload struct {
	file   *os.File
	offset uint64
	length uint32
}

type BlobStore struct {
	walf    *os.File
	WAL     *WAL
	buckets *xmaps.ConcurrentMap[string, *xmaps.ConcurrentMap[string, payload]]
}

func Open(path string) (*BlobStore, error) {
	store := &BlobStore{
		buckets: xmaps.NewConcurrentMap[string, *xmaps.ConcurrentMap[string, payload]](),
	}
	walFname := filepath.Join(path, "sdb.0.WAL")
	walf, err := os.OpenFile(walFname, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	WAL, err := NewWAL(walf, func(entry *Node) {
		bucketName := string(entry.bucket)
		keyName := string(entry.key)
		bucket, ok := store.buckets.Load(bucketName)
		if !ok {
			bucket = xmaps.NewConcurrentMap[string, payload]()
			store.buckets.Store(bucketName, bucket)
		}

		bucket.Store(keyName, payload{
			file:   walf,
			offset: 0, // TODO
			length: 0, // TODO
		})
	})

	if err != nil {
		return nil, err
	}

	store.WAL = WAL
	store.walf = walf
	return store, nil
}

func (b *BlobStore) Put(bucket, key string, value []byte) error {
	_, err := b.WAL.write(Node{
		kind:   walEntrySet,
		tx:     b.WAL.tx.Add(1),
		bucket: []byte(bucket),
		key:    []byte(key),
		val:    value,
	})

	return err
}

func (b *BlobStore) Get(bucketName, key string) ([]byte, error) {
	bucket, ok := b.buckets.Load(bucketName)
	if !ok {
		return nil, nil
	}

	_ = bucket
	return nil, nil
}

func (b *BlobStore) Close() error {
	return b.WAL.Close()
}
*/
