// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package fs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/tdb"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var _ blob.Store = (*BlobStore)(nil)

type dataInfo struct {
	ReferenceCount int64 `json:"c"`
}

type inode struct {
	Sha512 string `json:"h"` // sha512-256 as hex, which is way shorter than the default json int-array encoding
}

// BlobStore provides a file based implementation using a central metadata index db and blob deduplication with
// a two level fan-out directory structure.
// This lets the store scale into hundreds of millions of files, supports arbitrary key names and supports efficient
// range queries on names.
// The file layout is as follows:
//
//		baseDir
//		 |- _idx // directory with meta tdb
//		 +- ab                      // first byte of file hash, thus 255 folders at level 1
//		     +- 2f                 // second byte of file hash, thus 255 folders at level 2
//	           +- ax3ad3412123  // the actual data file, rest of sha512-256 hex-hash, 255*255*10.000 = 650mio files
//
// Because of the central index, it is more important than ever, to not use multiple instances on the same directory,
// because the index will be (logically) trashed and overwritten by each other.
//
// Note, that Windows is not officially supported and this store may eat your data or cause other corruptions.
// This is also a hard WON'T FIX, because we cannot delete open files properly (as defined by POSIX).
type BlobStore struct {
	baseDir         string
	dirLock         sync.RWMutex // today this is redundant, but when changing the bbolt store, this may get important
	db              *tdb.DB
	storePathIndex  *tdb.BlobStore // path -> hash
	storeRefCounts  *tdb.BlobStore // hash -> reference counter
	bucketNamePaths string
	bucketNameRC    string
	name            string
}

func NewBlobStore(baseDir string) (*BlobStore, error) {
	_ = os.MkdirAll(baseDir, os.ModePerm) // convenience to avoid bad file descriptor
	db, err := tdb.Open(filepath.Join(baseDir, "_idx"))
	if err != nil {
		return nil, err
	}

	b := &BlobStore{
		db:              db,
		baseDir:         baseDir,
		storePathIndex:  tdb.NewBlobStore(db, "p"),
		storeRefCounts:  tdb.NewBlobStore(db, "c"),
		bucketNameRC:    "c",
		bucketNamePaths: "p",
		name:            path.Base(baseDir),
	}

	return b, nil
}

func (b *BlobStore) Name() string {
	return b.name
}

func (b *BlobStore) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	return b.storePathIndex.List(ctx, opts)
}

func (b *BlobStore) Exists(ctx context.Context, key string) (bool, error) {
	return b.storePathIndex.Exists(ctx, key)
}

func (b *BlobStore) Close() error {
	return b.db.Close()
}

func (b *BlobStore) Delete(ctx context.Context, key string) error {
	b.dirLock.Lock()
	defer b.dirLock.Unlock()

	inodeBufOpt := b.db.Get(b.bucketNamePaths, key)
	if inodeBufOpt.IsNone() {
		// no such entry exists, that is fine
		return nil
	}

	inodeReader := inodeBufOpt.Unwrap()
	dec := json.NewDecoder(inodeReader)
	defer inodeReader.Close()

	var ind inode
	if err := dec.Decode(&ind); err != nil {
		// oops, some meta data error
		return fmt.Errorf("cannot parse inode meta data: %w", err)
	}

	if err := b.db.Delete(b.bucketNamePaths, key); err != nil {
		return err
	}

	ifoKey := string(ind.Sha512[:])
	dataInfoBufOpt := b.db.Get(b.bucketNameRC, ifoKey)
	if dataInfoBufOpt.IsNone() {
		return fmt.Errorf("meta data corrupted: inode exists but no inverse data info entry")
	}

	dataInfoReader := dataInfoBufOpt.Unwrap()
	dec = json.NewDecoder(dataInfoReader)
	defer dataInfoReader.Close()

	var ifo dataInfo
	if err := dec.Decode(&ifo); err != nil {
		return fmt.Errorf("cannot parse data info meta data: %w", err)
	}

	ifo.ReferenceCount--

	if ifo.ReferenceCount <= 0 {
		if err := b.db.Delete(b.bucketNameRC, ifoKey); err != nil {
			return fmt.Errorf("rc dropped to 0, but cannot delete reference count from db: %w", err)
		}

		// this was the last reference, thus remove it also from the filesystem
		fname := b.filepath(ind.Sha512)
		err := os.Remove(fname)
		if err != nil && !os.IsNotExist(err) {
			// this may be a permission problem or just running on windows and having a Reader open...
			return fmt.Errorf("rc dropped to 0, but cannot delete physical data file: %w", err)
		}
	} else {
		// just lost one rc, thus just persist the new count
		buf, err := json.Marshal(ifo)
		if err != nil {
			panic(fmt.Errorf("cannot happen: error on marshalling data info entry: %w", err))
		}

		if err := b.db.Set(b.bucketNameRC, ifoKey, buf); err != nil {
			return fmt.Errorf("failed to update data info entry: %w", err)
		}
	}

	return nil
}

func (b *BlobStore) filepath(hash string) string {
	h := hash
	fan0 := h[:2]
	h = h[2:]
	fan1 := h[:2]
	h = h[2:]
	return filepath.Join(b.baseDir, fan0, fan1, h)
}

func (b *BlobStore) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	var fname string

	b.dirLock.RLock()
	defer b.dirLock.RUnlock()

	inodeBufOpt := b.db.Get(b.bucketNamePaths, key)
	if inodeBufOpt.IsNone() {
		return std.None[io.ReadCloser](), nil
	}

	reader := inodeBufOpt.Unwrap()
	dec := json.NewDecoder(reader)
	defer reader.Close()

	var ind inode
	if err := dec.Decode(&ind); err != nil {
		// oops, some meta data error
		return std.None[io.ReadCloser](), fmt.Errorf("cannot parse inode meta data: %w", err)
	}

	fname = b.filepath(ind.Sha512)

	var res std.Option[io.ReadCloser]

	if fname == "" {
		return res, nil
	}

	// here we rely on the unix philosophy: deleting open files is fine
	f, err := os.Open(fname)
	if err != nil {
		if !os.IsNotExist(err) {
			return res, nil
		}

		return res, fmt.Errorf("cannot open physical data file: %w", err)
	}

	return std.Some[io.ReadCloser](f), nil
}

// NewWriter first freely writes into a temporary file and afterward will only block as short as possible
// to update any metadata to avoid our dreaded deadlocks. The commit happens when closing the writer.
func (b *BlobStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	// allocate a temporary file
	var rbuf [16]byte
	if _, err := rand.Read(rbuf[:]); err != nil {
		return nil, fmt.Errorf("cannot get random bytes: %w", err)
	}

	tmpFname := filepath.Join(b.baseDir, hex.EncodeToString(rbuf[:])+".tmp")
	tmpf, err := os.OpenFile(tmpFname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file '%s': %w", tmpFname, err)
	}

	return &txWriter{key: key, ctx: ctx, parent: b, File: tmpf, tmpFname: tmpFname}, nil
}
