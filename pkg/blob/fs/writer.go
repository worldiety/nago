package fs

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
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
	copy(ind.Sha512[:], hash)
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

	// we may consider a fsync here at directory level, but MacBooks ignore that and fileservers have USV, so
	// at worst we will later have a missing physical blob.

	// target file is physically at place, now update our db. If that doesn't work, we have at worst a stale file,
	// not that bad, just a few wasted bytes but still consistent.
	err = t.parent.db.Update(func(tx *bbolt.Tx) error {
		t.parent.dirLock.Lock()
		defer t.parent.dirLock.Unlock()

		// write the inode
		indBuf, err := json.Marshal(ind)
		if err != nil {
			return fmt.Errorf("cannot marshal inode: %w", err)
		}

		bucket := tx.Bucket(t.parent.bucketNamePaths)
		if bucket == nil {
			bucket, err = tx.CreateBucketIfNotExists(t.parent.bucketNamePaths)
			if err != nil {
				return fmt.Errorf("cannot create inode bucket: %w", err)
			}
		}

		if err := bucket.Put([]byte(t.key), indBuf); err != nil {
			return fmt.Errorf("cannot put inode into bucket: %w", err)
		}

		// write the rc
		rcBucket := tx.Bucket(t.parent.bucketNameRC)
		if rcBucket == nil {
			rcBucket, err = tx.CreateBucketIfNotExists(t.parent.bucketNameRC)
			if err != nil {
				return fmt.Errorf("cannot create rc bucket: %w", err)
			}
		}

		var ifo dataInfo
		dataInfoBuf := rcBucket.Get(ind.Sha512[:])
		if dataInfoBuf != nil {
			if err := json.Unmarshal(dataInfoBuf, &ifo); err != nil {
				return fmt.Errorf("cannot unmarshal dataInfo from bucket: %w", err)
			}
		}

		ifo.ReferenceCount++

		dataInfoBuf, err = json.Marshal(&ifo)
		if err != nil {
			return fmt.Errorf("cannot marshal dataInfo: %w", err)
		}

		if err := rcBucket.Put(ind.Sha512[:], dataInfoBuf); err != nil {
			return fmt.Errorf("cannot put dataInfo to bucket: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot update meta data for newly written file '%s': %w", t.tmpFname, err)
	}

	t.closed = true

	return nil
}
