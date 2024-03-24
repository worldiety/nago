package fs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BlobStore provides a file based implementation without transactions.
// The transactions are just fake implementations to satisfy the contract and respect the read/write property.
// However, the store itself is at least thread safe in the same way the filesystem is.
// This implementation uses a flat hex projection of the key into a single directory, so keep your amount
// of entries below 10.000 otherwise your system performance will degrade. Do not share folders concurrently.
// The empty key is not allowed and due to the hex encoding (to guarantee support for all utf-8 strings
// on any filesystem) the filename length is doubled. This also protects against path traversal attacks.
type BlobStore struct {
	baseDir string
}

func NewBlobStore(baseDir string) *BlobStore {
	_ = os.MkdirAll(baseDir, os.ModePerm) // convenience to avoid bad file descriptor
	return &BlobStore{baseDir: baseDir}
}

func (b *BlobStore) Update(f func(blob.Tx) error) error {
	tx := &fsTx{parent: b}
	if err := f(tx); err != nil {
		return err
	}

	return nil
}

func (b *BlobStore) View(f func(blob.Tx) error) error {
	tx := &fsTx{parent: b}
	if err := f(tx); err != nil {
		return err
	}

	return nil
}

func (b *BlobStore) fname(key string) string {
	return filepath.Join(b.baseDir, hex.EncodeToString([]byte(key))) + ".blob"
}

type fsTx struct {
	parent   *BlobStore
	readOnly bool
}

func (f *fsTx) Each(yield func(blob.Entry, error) bool) {
	files, err := os.ReadDir(f.parent.baseDir)
	if err != nil {
		yield(blob.Entry{}, err)
		return
	}

	for _, file := range files {
		if !file.Type().IsRegular() || !strings.HasSuffix(file.Name(), ".blob") {
			continue
		}

		fname := filepath.Join(f.parent.baseDir, file.Name())
		buf, err := hex.DecodeString(file.Name()[:len(file.Name())-5])
		yield(blob.Entry{
			Key: string(buf),
			Open: func() (io.ReadCloser, error) {
				if err != nil {
					return nil, err
				}
				return os.Open(fname) // intentionally read-only
			},
		}, err)
	}

}

func (f *fsTx) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("fs implementation does not allow empty key")
	}
	err := os.Remove(f.parent.fname(key))
	if os.IsNotExist(err) {
		return nil
	}

	return err // nil or any other like permission
}

// Put is not transactional, however we apply unix atomic rename semantics for each individual file.
// Thus, interrupted or crashing processes cannot corrupt any existing or partially written files.
func (f *fsTx) Put(entry blob.Entry) error {
	if entry.Key == "" {
		return fmt.Errorf("fs implementation does not allow empty key")
	}

	src, err := entry.Open()
	if err != nil {
		return fmt.Errorf("cannot open entry stream: %w", err)
	}

	fname := f.parent.fname(entry.Key)
	var rbuf [16]byte
	if _, err := rand.Read(rbuf[:]); err != nil {
		return fmt.Errorf("cannot get random bytes: %w", err)
	}

	tmpFname := hex.EncodeToString(rbuf[:]) + ".tmp"
	tmp, err := os.OpenFile(tmpFname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot create temp file '%s': %w", tmpFname, err)
	}

	if _, err := io.Copy(tmp, src); err != nil {
		if err := tmp.Close(); err != nil {
			_ = err // suppressed error, primary error is the copy failure
		}

		_ = os.Remove(tmpFname) // don't care either, but try to cleanup

		return fmt.Errorf("cannot copy bytes: %w", err)
	}

	// note that many vfs (e.g. fuse) will delay writes until close
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot complete and close tmp file '%s': %w", tmpFname, err)
	}

	// finally perform a unix atomic file replacement which removes the dst file by definition
	return os.Rename(tmpFname, fname)
}

func (f *fsTx) Get(key string) (std.Option[blob.Entry], error) {
	if key == "" {
		return std.None[blob.Entry](), fmt.Errorf("fs implementation does not allow empty key")
	}
	fname := f.parent.fname(key)

	return std.Some(blob.Entry{
		Key: key,
		Open: func() (io.ReadCloser, error) {
			file, err := os.Open(fname) // intentionally read-only
			if err != nil {
				return nil, fmt.Errorf("cannot open '%s': %w", fname, err)
			}

			return file, nil
		},
	}), nil
}
