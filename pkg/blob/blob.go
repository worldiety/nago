package blob

import (
	"bytes"
	"go.wdy.de/nago/pkg/std"
	"io"
)

// Store abstraction for generic binary streams.
type Store interface {
	// Update executes a read-write transaction. This may block any other write and read transactions
	// (implementation dependent).
	Update(f func(Tx) error) error
	// View executes a read-only transaction, which allows usually higher concurrency optimizations
	// (implementation dependent).
	View(f func(Tx) error) error
}

// Entry represents either a new or an existing entry and is usually only valid within its scoping
// and pending transaction. This allows to optimize iteration of entries based on keys.
type Entry struct {
	Key  string
	Open func() (io.ReadCloser, error) // caller must close within the lifetime of transaction
}

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
