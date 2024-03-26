package application

import (
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/internal/text"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/persistence/kv/bolt"
	"go.wdy.de/nago/pkg/blob"
	bolt2 "go.wdy.de/nago/pkg/blob/bolt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"os"
	"path/filepath"
	"reflect"
)

// deprecated: use BlobStore
//
// Store returns a configured transactional key value store by name
// or panics and switches into maintenance mode.
func (c *Configurator) Store(name string) kv.Store {
	if c.appName == "" {
		panic("set app name first")
	}

	if store, ok := c.kvStores[name]; ok {
		return store
	}

	dir, _ := os.Getwd()
	dir = filepath.Join(dir, "."+text.SafeName(c.appName), "kvstore")
	_ = os.MkdirAll(dir, os.ModePerm)
	fname := filepath.Join(dir, text.SafeName(name)+".db")
	db, err := bbolt.Open(fname, os.ModePerm, nil)
	if err != nil {
		panic(fmt.Errorf("cannot open bbolt database '%s': %w", fname, err))
	}

	store := bolt.NewStore(db)
	c.kvStores[name] = store
	c.boltStores[name] = db

	return store
}

// BlobStore creates a new blob store instance, currently a bbolt implementation.
// deprecated: don't know if this is a good thing. See [Repository].
func (c *Configurator) BlobStore(dbName, bucketName string) blob.Store {
	c.Store(dbName)
	db := c.boltStores[dbName]
	return bolt2.NewBlobStore(db, bucketName)
}

// SloppyRepository returns a default Repository implementation for the given type, which just serializes the domain
// type, which is fine for rapid prototyping, but should not be used for products which must be maintained.
func SloppyRepository[A data.Aggregate[ID], ID data.IDType](cfg *Configurator) data.Repository[A, ID] {
	var zero A
	bucketName := reflect.TypeOf(zero).Name()
	store := cfg.BlobStore("nago.db", bucketName)
	return json.NewSloppyJSONRepository[A, ID](store)
}
