package application

import (
	"fmt"
	badger2 "github.com/dgraph-io/badger/v4"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/badger"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
)

// EntityStore returns the default applications blob store. There is only one instance.
// Do not put (large) files into this store. See also [blob.Get] and [blob.Put] helper functions.
// For just storing serialized repository data, consider using [SloppyRepository] or
// a [json.NewJSONRepository] with custom domain model mapping.
//
// Use [Configurator.FileStore] for large blobs.
func (c *Configurator) EntityStore(bucketName string) blob.Store {
	store, ok := c.stores[bucketName]
	if ok {
		return store
	}

	if c.globalBadger == nil {
		requiresBolt2BadgerMigration := !c.hasBadger() && c.hasBolt()

		db, err := badger.Open(c.Directory("badgerdb"))
		if err != nil {
			panic(fmt.Errorf("could not open badger store: %v", err))
		}

		c.globalBadger = db.Unwrap()

		if requiresBolt2BadgerMigration {
			if err := c.migrateBBolt2Badger(); err != nil {
				panic(err)
			}
		}
	}

	db := badger.NewBlobStore(c.globalBadger)
	db.SetPrefix(bucketName)
	c.stores[bucketName] = db

	return db
}

func (c *Configurator) hasBadger() bool {
	dir := c.directory("badgerdb")
	files, _ := os.ReadDir(dir)
	return len(files) > 0
}

func (c *Configurator) hasBolt() bool {
	fname := filepath.Join(c.Directory("bbolt"), "bolt.db")
	_, err := os.Stat(fname)
	return !os.IsNotExist(err)
}

func (c *Configurator) migrateBBolt2Badger() error {
	slog.Warn("detected bolt store, migrating to badgerdb")
	fname := filepath.Join(c.Directory("bbolt"), "bolt.db")
	db, err := bbolt.Open(fname, 0700, nil) // security: only owner can read,write,exec
	if err != nil {
		return fmt.Errorf("cannot open bbolt database '%s': %w", fname, err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		return c.globalBadger.Update(func(txn *badger2.Txn) error {
			return tx.ForEach(func(bucketName []byte, b *bbolt.Bucket) error {
				return b.ForEach(func(k, v []byte) error {
					slog.Info("migrating", "bucket", string(bucketName), "key", string(k))
					prefixedKey := append(bucketName, k...)
					return txn.Set(prefixedKey, v)
				})
			})

		})
	})

	if err != nil {
		return err
	}

	return db.Sync()

}

// FileStore returns a blob store which directly saves into the filesystem and is recommended for handling large
// files. See also [blob.Read] and [blob.Write] helper functions.
func (c *Configurator) FileStore(bucketName string) blob.Store {
	dir := c.Directory(filepath.Join("files", bucketName))
	slog.Info(fmt.Sprintf("file store '%s' stores in '%s'", bucketName, dir))
	store, err := fs.NewBlobStore(dir)
	if err != nil {
		panic(fmt.Errorf("cannot open file blob store '%s': %w", dir, err))
	}

	return store
}

// SloppyRepository returns a default Repository implementation for the given type, which just serializes the domain
// type, which is fine for rapid prototyping, but should not be used for products which must be maintained.
// This shares the bucket name space with [Configurator.EntityStore] and uses the reflected type name as the
// bucket name, so be careful when renaming types or having type name collisions.
func SloppyRepository[A data.Aggregate[ID], ID data.IDType](cfg *Configurator) data.Repository[A, ID] {
	var zero A
	bucketName := reflect.TypeOf(zero).Name()
	store := cfg.EntityStore(bucketName)
	return json.NewSloppyJSONRepository[A, ID](store)
}
