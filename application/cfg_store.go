package application

import (
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/internal/text"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/persistence/kv/bolt"
	"os"
	"path/filepath"
)

// Store returns a configured transactional key value store by name
// or panics and switches into maintenance mode.
func (c *Configurator) Store(name string) kv.Store {
	if c.appName == "" {
		panic("set app name first")
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
	return store
}
