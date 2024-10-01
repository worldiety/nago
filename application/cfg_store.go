package application

import (
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/blob/tdb"
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

	if c.globalTDB == nil {
		requiresBolt2TDBMigration := !c.hasTDB() && c.hasBolt()

		db, err := tdb.Open(c.Directory("tdb"))
		if err != nil {
			panic(fmt.Errorf("could not open tdb store: %v", err))
		}

		c.globalTDB = db

		if requiresBolt2TDBMigration {
			if err := c.migrateBBolt2TDB(); err != nil {
				panic(err)
			}
		}
	}

	db := tdb.NewBlobStore(c.globalTDB, bucketName)
	c.stores[bucketName] = db

	return db
}

func (c *Configurator) hasTDB() bool {
	dir := c.directory("tdb")
	files, _ := os.ReadDir(dir)
	return len(files) > 0
}

func (c *Configurator) hasBolt() bool {
	fname := filepath.Join(c.Directory("bbolt"), "bolt.db")
	_, err := os.Stat(fname)
	return !os.IsNotExist(err)
}

func (c *Configurator) migrateBBolt2TDB() error {
	slog.Warn("detected bolt store, migrating to tdb")
	fname := filepath.Join(c.Directory("bbolt"), "bolt.db")
	db, err := bbolt.Open(fname, 0700, nil) // security: only owner can read,write,exec
	if err != nil {
		return fmt.Errorf("cannot open bbolt database '%s': %w", fname, err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucketName []byte, b *bbolt.Bucket) error {
			buckName := string(bucketName)
			return b.ForEach(func(k, v []byte) error {
				slog.Info("migrating", "bucket", string(bucketName), "key", string(k))
				return c.globalTDB.Set(buckName, string(k), v)
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
