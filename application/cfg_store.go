package application

import (
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob"
	bolt2 "go.wdy.de/nago/pkg/blob/bolt"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"log/slog"
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
	if c.boltStore == nil {
		fname := filepath.Join(c.Directory("bbolt"), "bolt.db")
		db, err := bbolt.Open(fname, 0700, nil) // security: only owner can read,write,exec
		if err != nil {
			panic(fmt.Errorf("cannot open bbolt database '%s': %w", fname, err))
		}

		c.boltStore = db
		slog.Info("bbolt store opened", slog.String("file", fname))
	}

	slog.Info("BlobStore bucket opened", slog.String("bucket", bucketName), slog.String("file", c.boltStore.Path()))

	return bolt2.NewBlobStore(c.boltStore, bucketName)
}

// FileStore returns a blob store which directly saves into the filesystem and is recommended for handling large
// files. See also [blob.Read] and [blob.Write] helper functions.
func (c *Configurator) FileStore(bucketName string) blob.Store {
	dir := c.Directory(filepath.Join("files", bucketName))
	slog.Info(fmt.Sprintf("file store '%s' stores in '%s'", bucketName, dir))
	return fs.NewBlobStore(dir)
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
