// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/blob/tdb"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sync"
)

// EntityStore returns the default applications blob store. There is only one instance.
// Do not put (large) files into this store. See also [blob.Get] and [blob.Put] helper functions.
// For just storing serialized repository data, consider using [SloppyRepository] or
// a [json.NewJSONRepository] with custom domain model mapping.
//
// Use [Configurator.FileStore] for large blobs.
func (c *Configurator) EntityStore(bucketName string) (blob.Store, error) {
	stores, err := c.Stores()
	if err != nil {
		return nil, err
	}

	return stores.Open(bucketName, blob.OpenStoreOptions{Type: blob.EntityStore})
}

// FileStore returns a blob store which directly saves into the filesystem and is recommended for handling large
// files. See also [blob.Read] and [blob.Write] helper functions.
func (c *Configurator) FileStore(bucketName string) (blob.Store, error) {
	stores, err := c.Stores()
	if err != nil {
		return nil, err
	}

	return stores.Open(bucketName, blob.OpenStoreOptions{Type: blob.FileStore})
}

// Stores tries to open the local filesystem writeable and allocates the space for entity stores.
func (c *Configurator) Stores() (blob.Stores, error) {
	c.storesMutex.Lock()
	defer c.storesMutex.Unlock()

	if c.stores == nil {
		stores, err := NewLocalStores(c.DataDir())
		if err != nil {
			return nil, err
		}

		c.stores = stores
	}

	return c.stores, nil
}

// Deprecated: use JSONRepository
//
// SloppyRepository returns a default Repository implementation for the given type, which just serializes the domain
// type, which is fine for rapid prototyping, but should not be used for products which must be maintained.
// This shares the bucket name space with [Configurator.EntityStore] and uses the reflected type name as the
// bucket name, so be careful when renaming types or having type name collisions.
func SloppyRepository[A data.Aggregate[ID], ID data.IDType](cfg *Configurator) data.Repository[A, ID] {
	bucketName := reflect.TypeFor[A]().Name()
	return JSONRepository[A, ID](cfg, bucketName)
}

// JSONRepository returns a sloppy json Repository implementation for the given type, which just serializes the domain
// type, which is fine for rapid prototyping, but should be used with care for products which must be maintained.
// This shares the bucket name space with [Configurator.EntityStore].
func JSONRepository[A data.Aggregate[ID], ID data.IDType](cfg *Configurator, bucketName string) data.Repository[A, ID] {
	store := std.Must(cfg.EntityStore(bucketName))
	return json.NewSloppyJSONRepository[A, ID](store)
}

// security: only owner can read,write,exec
const defaultDirPermission = 0700

type storeEntry struct {
	store blob.Store
	info  blob.StoreInfo
}
type LocalStores struct {
	stores            concurrent.RWMap[string, *storeEntry]
	mutex             sync.Mutex
	db                *tdb.DB
	tdbRootDir        string
	blobBucketRootDir string
}

func NewLocalStores(rootDir string) (*LocalStores, error) {
	tdbRoot := filepath.Join(rootDir, "tdb")
	if err := os.MkdirAll(tdbRoot, defaultDirPermission); err != nil {
		return nil, fmt.Errorf("cannot create root folder for tdb: %w", err)
	}

	db, err := tdb.Open(tdbRoot)
	if err != nil {
		return nil, fmt.Errorf("could not open tdb store: %v", err)
	}

	blobBucketRootDir := filepath.Join(rootDir, "files")
	if err := os.MkdirAll(blobBucketRootDir, defaultDirPermission); err != nil {
		return nil, fmt.Errorf("cannot create root folder for tdb: %w", err)
	}

	return &LocalStores{
		db:                db,
		tdbRootDir:        tdbRoot,
		blobBucketRootDir: blobBucketRootDir,
	}, nil
}

func (s *LocalStores) All() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for name := range s.stores.All() {
			if !yield(name, nil) {
				return
			}
		}
	}
}

func (s *LocalStores) Stat(name string) (option.Opt[blob.StoreInfo], error) {
	entry, ok := s.stores.Get(name)
	if !ok {
		return option.None[blob.StoreInfo](), nil
	}

	return option.Some(entry.info), nil
}

// Get returns any known store and may open it, if the implementation knows the type.
func (s *LocalStores) Get(name string) (option.Opt[blob.Store], error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entry, ok := s.stores.Get(name)
	if ok {
		return option.Some(entry.store), nil
	}

	// we never opened it, thus inspect, if it may be an entity store
	for n := range s.db.Buckets() {
		if n == name {
			store, err := s.open(name, blob.OpenStoreOptions{Type: blob.EntityStore})
			if err != nil {
				return option.None[blob.Store](), fmt.Errorf("cannot open entity store %s: %w", name, err)
			}

			return option.Some(store), nil
		}
	}

	// try to peek into file stores
	dir := filepath.Join(s.blobBucketRootDir, name)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return option.None[blob.Store](), nil
		}

		return option.None[blob.Store](), err
	}

	if len(files) == 0 {
		return option.None[blob.Store](), nil
	}

	// looks like there is some data
	store, err := s.open(name, blob.OpenStoreOptions{Type: blob.FileStore})
	if err != nil {
		return option.None[blob.Store](), fmt.Errorf("cannot open file store %s: %w", name, err)
	}

	return option.Some(store), nil
}

func (s *LocalStores) Open(name string, opts blob.OpenStoreOptions) (blob.Store, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.open(name, opts)
}

func (s *LocalStores) open(name string, opts blob.OpenStoreOptions) (blob.Store, error) {
	if err := validateStoreName(name); err != nil {
		return nil, err
	}

	entry, ok := s.stores.Get(name)
	if ok {
		if entry.info.Type != opts.Type {
			return nil, fmt.Errorf("type mismatch of already opened store '%s': %v", name, entry.info.Type)
		}

		return entry.store, nil
	}

	switch opts.Type {
	case blob.FileStore:
		dir := filepath.Join(s.blobBucketRootDir, name)
		if err := os.MkdirAll(dir, defaultDirPermission); err != nil {
			return nil, fmt.Errorf("cannot create root folder for tdb: %w", err)
		}

		store, err := fs.NewBlobStore(dir)
		if err != nil {
			return nil, fmt.Errorf("cannot open file blob store '%s': %w", dir, err)
		}

		entry = &storeEntry{
			store: store,
			info: blob.StoreInfo{
				Type: opts.Type,
			},
		}
	case blob.EntityStore:
		db := tdb.NewBlobStore(s.db, name)
		entry = &storeEntry{
			store: db,
			info: blob.StoreInfo{
				Type: opts.Type,
			},
		}

	default:
		return nil, fmt.Errorf("unsupported store type: %v", opts.Type)
	}

	s.stores.Put(name, entry)

	return entry.store, nil
}

func (s *LocalStores) SetContentTypes(name string, types []blob.ContentType) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	info, ok := s.stores.Get(name)
	if !ok {
		return fmt.Errorf("store '%s' not found", name)
	}

	info.info.ContentTypes = xslices.New(types...)
	return nil
}

func (s *LocalStores) Delete(name string) error {
	// tdb cannot delete and there are other races with the fs stores while store is opened
	return fmt.Errorf("deletion of stores is not yet implemented")
}

// note that we allowed uppercase letters in the past and thus we must still allow them, but that was a bad oversight.
var regexStoreName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func validateStoreName(name string) error {
	if !regexStoreName.MatchString(name) {
		return fmt.Errorf("invalid store name: %s", name)
	}

	return nil
}
