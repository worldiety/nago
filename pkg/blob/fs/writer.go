// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package fs

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type txWriter struct {
	key      string
	parent   *BlobStore
	tmpFname string
	ctx      context.Context
	*os.File
	closed bool
}

func (t *txWriter) Close() error {
	if t.closed {
		return nil
	}

	if t.ctx.Err() != nil {
		return t.ctx.Err()
	}

	// note that many vfs (e.g. fuse or NFS?) will delay writes until close
	if err := t.File.Close(); err != nil {
		return fmt.Errorf("cannot complete and close tmp file '%s': %w", t.tmpFname, err)
	}

	// there are so many optimized writer interfaces, which we inherit (like io.WriterAt), thus we
	// cannot just update a hasher on each write call: we must read the file from scratch
	hasher := sha512.New512_256()
	tmpF, err := os.Open(t.tmpFname)
	if err != nil {
		return fmt.Errorf("cannot open tmp file for hashing '%s': %w", t.tmpFname, err)
	}

	if _, err := io.Copy(hasher, tmpF); err != nil {
		return fmt.Errorf("cannot hash tmp file '%s': %w", t.tmpFname, err)
	}

	hash := hasher.Sum(nil)
	var ind inode
	ind.Sha512 = hex.EncodeToString(hash)
	targetFname := t.parent.filepath(ind.Sha512)

	// ensure fan-out
	if err := os.MkdirAll(filepath.Dir(targetFname), 0755); err != nil {
		return fmt.Errorf("cannot create target fan-out directory '%s': %w", t.parent.filepath(ind.Sha512), err)
	}

	// perform a unix atomic file replacement which removes the dst file by definition.
	// this automatically deduplicates, as it unlinks the destination (which contains the same data).
	if err := os.Rename(t.tmpFname, targetFname); err != nil {
		return fmt.Errorf("cannot rename tmp file '%s' to '%s': %w", t.tmpFname, targetFname, err)
	}

	// we may consider a fsync here at directory level, but MacBooks are slow and fileservers have USV, so
	// at worst we will later have a missing physical blob.

	// target file is physically at place, now update our db. If that doesn't work, we have at worst a stale file,
	// not that bad, just a few wasted bytes but still consistent.
	t.parent.dirLock.Lock()
	defer t.parent.dirLock.Unlock()

	// write the inode
	indBuf, err := json.Marshal(ind)
	if err != nil {
		return fmt.Errorf("cannot marshal inode: %w", err)
	}

	if err := t.parent.db.Set(t.parent.bucketNamePaths, t.key, indBuf); err != nil {
		return fmt.Errorf("cannot put inode into bucket: %w", err)
	}

	// write the rc
	ifoKey := string(ind.Sha512[:])

	var ifo dataInfo
	dataInfoBufOpt := t.parent.db.Get(t.parent.bucketNameRC, ifoKey)
	if dataInfoBufOpt.IsSome() {
		reader := dataInfoBufOpt.Unwrap()
		defer reader.Close()
		dec := json.NewDecoder(reader)

		if err := dec.Decode(&ifo); err != nil {
			return fmt.Errorf("cannot unmarshal dataInfo from bucket: %w", err)
		}
	}

	ifo.ReferenceCount++

	dataInfoBuf, err := json.Marshal(&ifo)
	if err != nil {
		return fmt.Errorf("cannot marshal dataInfo: %w", err)
	}

	if err := t.parent.db.Set(t.parent.bucketNameRC, ifoKey, dataInfoBuf); err != nil {
		return fmt.Errorf("cannot put dataInfo to bucket: %w", err)
	}

	t.closed = true

	return nil
}
