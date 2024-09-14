package blob

import (
	"bytes"
	"context"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
)

// Deprecated: the implications cause scale issues and deadlock headaches
// Store abstraction for generic binary streams.
// TODO this transactional spec does not work properly:
//   - bbolt cannot nest transactions without deadlocks
//   - fs, s3 do not support transactions at all
//   - only universally supported by postgresql
//   - mariadb/mysql do not support transactions for schema changes (== creating and deleting buckets)
//   - mongodb, aws documentdb (one per session, thus not nested) not sure about GCP Firestore
//   - session logic will have a huge scaling problem, if required and applied in contrast to eventual and non transactional systems
type Store interface {
	// Update executes a read-write transaction. This may block any other write and read transactions
	// (implementation dependent).
	Update(f func(Tx) error) error
	// View executes a read-only transaction, which allows usually higher concurrency optimizations
	// (implementation dependent).
	View(f func(Tx) error) error
}

type ListOptions struct {
	// If non-zero, the result set will only contain keys which starts with the given prefix.
	Prefix string
	// If non-zero, MinInc marks the inclusive minimal starting key in the result set.
	MinInc string
	// If non-zero, MaxInc marks the inclusive minimal ending key in the result set.
	MaxInc string
}

// Store represents a single bucket store for blobs. Note, that individual methods are thread safe, however
// it is not possible to represent transactions.
// This limitation is intentionally, because neither simple implementations (an ordinary filesystem) nor
// scaling out implementations (eventual consistent clustered cloud storage) support proper transactions.
type Store2 interface {
	// List takes a snapshot of all available entries and returns an iterator for it.
	// While iterating, any operation on the dataset can be performed without blocking, however
	// these changes must not cause the iterator to return garbage (like missed or doubled entries).
	// Note that this may become very inefficient, when used on very large datasets containing
	// millions or even billions of entries. The order of the returned keys is implementation dependent.
	// Implementations must support Prefix and Range filters.
	List(ctx context.Context, opts ListOptions) iter.Seq2[string, error]

	// Exists returns only true, if at least at some time such blob existed. Note, that in concurrent situations
	// such a statement is not very useful.
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes the denoted entry. It is not an error to remove a non-existent file.
	Delete(ctx context.Context, key string) error

	// NewReader opens the blob to be read.
	NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error)

	// NewWriter open the blob to be created or overwritten. Either of them will only
	// if the writer has been closed and the context has not been cancelled.
	// A Write is always atomic and implementations must ensure, that
	// a partial write is never visible.
	NewWriter(ctx context.Context, key string) (io.WriteCloser, error)
}

type ListObjectsOptions struct {
	Prefix string
}

// Entry represents either a new or an existing entry and is usually only valid within its scoping
// and pending transaction. This allows to optimize iteration of entries based on keys.
type Entry struct {
	Key  string
	Open func() (io.ReadCloser, error) // caller must close within the lifetime of transaction
}

// Deprecated: the implications cause scale issues and deadlock headaches
//
// Tx is a transaction scope. Implementations without transaction support will just provide a fake Tx instance.
// It is generally not safe to nest transactions and may easily cause deadlocks, depending on the actual implementation.
type Tx interface {
	// Each loops over each entry and provides an open function.
	// [Entry.Open] is owned by the yield and only valid during the yield call.
	// Entry is only valid for the lifetime of the Tx.
	// This is a [iter.Seq2].
	Each(yield func(Entry, error) bool)
	// Delete removes an entry. it is not an error to delete a non-existing entry.
	Delete(key string) error
	// Put creates or updates the target entry.
	Put(entry Entry) error
	// Get returns the entry which is only valid for the lifetime of the enclosing Tx.
	Get(key string) (std.Option[Entry], error)
}

// Read transfers from the store all bytes into the given writer, e.g. into a http response.
func Read(store Store, key string, dst io.Writer) error {
	err := store.View(func(tx Tx) error {
		optEnt, err := tx.Get(key)
		if err != nil {
			return err
		}

		reader, err := optEnt.Unwrap().Open()
		if err != nil {
			return err
		}

		defer reader.Close()

		_, err = io.Copy(dst, reader)
		return err
	})

	return err
}

// Write transfers all bytes from the given source into the store, e.g. from a request body.
func Write(store Store, key string, src io.Reader) error {
	return store.Update(func(tx Tx) error {
		return tx.Put(Entry{
			Key: key,
			Open: func() (io.ReadCloser, error) {
				return readerCloser{Reader: src}, nil
			},
		})
	})
}

// Put is a shorthand function to write small values using a slice into the store. Do not use for large blobs.
func Put(store Store, key string, value []byte) error {
	return store.Update(func(tx Tx) error {
		return tx.Put(Entry{
			Key: key,
			Open: func() (io.ReadCloser, error) {
				return readerCloser{Reader: bytes.NewReader(value)}, nil
			},
		})
	})
}

func Delete(store Store, key string) error {
	return store.Update(func(tx Tx) error {
		return tx.Delete(key)
	})
}

func DeleteAll(store Store) error {
	return store.Update(func(tx Tx) error {
		var e error
		tx.Each(func(entry Entry, err error) bool {
			if err != nil {
				e = err
				return false
			}
			// TODO not sure if deleting while iterating should be well defined
			if err := tx.Delete(entry.Key); err != nil {
				e = err
				return false
			}

			return true
		})

		return e
	})
}

// Get is a shortcut function to read small slices from the store. Do not use for large blobs, because it allocates
// the entire blob size without other limits.
func Get(store Store, key string) (std.Option[[]byte], error) {
	var res std.Option[[]byte]
	err := store.View(func(tx Tx) error {
		optEnt, err := tx.Get(key)
		if err != nil {
			return err
		}

		if !optEnt.Valid {
			return nil
		}

		reader, err := optEnt.Unwrap().Open()
		if err != nil {
			return err
		}

		defer reader.Close()

		buf, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		res = std.Some(buf)
		return nil
	})

	return res, err
}

type readerCloser struct {
	io.Reader
}

func (readerCloser) Close() error {
	return nil
}

func Keys(store Store) ([]string, error) {
	var keys []string
	err := store.View(func(tx Tx) error {
		var e error
		tx.Each(func(entry Entry, err error) bool {
			if err != nil {
				e = err
				return false
			}

			keys = append(keys, entry.Key)
			return true
		})
		return e
	})

	return keys, err
}
