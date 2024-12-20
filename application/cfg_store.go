package application

import (
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/blob/tdb"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
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
func (c *Configurator) EntityStore(bucketName string) (blob.Store, error) {
	store, ok := c.entityStores[bucketName]
	if ok {
		return store, nil
	}

	if c.globalTDB == nil {

		db, err := tdb.Open(c.Directory("tdb"))
		if err != nil {
			return nil, fmt.Errorf("could not open tdb store: %v", err)
		}

		c.globalTDB = db

	}

	db := tdb.NewBlobStore(c.globalTDB, bucketName)
	c.entityStores[bucketName] = db

	return db, nil
}

// FileStore returns a blob store which directly saves into the filesystem and is recommended for handling large
// files. See also [blob.Read] and [blob.Write] helper functions.
func (c *Configurator) FileStore(bucketName string) (blob.Store, error) {
	store, ok := c.fileStores[bucketName]
	if ok {
		return store, nil
	}

	dir := c.Directory(filepath.Join("files", bucketName))
	slog.Info(fmt.Sprintf("file store '%s' stores in '%s'", bucketName, dir))
	store, err := fs.NewBlobStore(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot open file blob store '%s': %w", dir, err)
	}

	c.fileStores[bucketName] = store

	return store, nil
}

// SloppyRepository returns a default Repository implementation for the given type, which just serializes the domain
// type, which is fine for rapid prototyping, but should not be used for products which must be maintained.
// This shares the bucket name space with [Configurator.EntityStore] and uses the reflected type name as the
// bucket name, so be careful when renaming types or having type name collisions.
func SloppyRepository[A data.Aggregate[ID], ID data.IDType](cfg *Configurator) data.Repository[A, ID] {
	var zero A
	bucketName := reflect.TypeOf(zero).Name()
	store := std.Must(cfg.EntityStore(bucketName))
	return json.NewSloppyJSONRepository[A, ID](store)
}
