package blob

import (
	"go.wdy.de/nago/pkg/std"
	"io"
)

type Store interface {
	Update(f func(Tx) error) error
	View(f func(Tx) error) error
}

type Entry struct {
	Key  string
	Open func() (io.ReadCloser, error) // caller must close within the lifetime of transaction
}

// Tx is a transaction scope. Implementations without transaction support will just provide a fake Tx instance.
// It is generally not safe to nest transactions and may easily cause deadlocks, depending on the actual implementation.
type Tx interface {
	// Each loops over each entry and provides an open function.
	// The key and value are owned by the Tx and must not be kept alive outside of the according yield call,
	// because they may get re-used.
	// This is a [iter.Seq2].
	Each(yield func(Entry, error) bool)
	// Delete removes an entry. it is not an error to delete a non-existing entry.
	Delete(key string) error
	// Put creates or updates the target entry.
	Put(entry Entry) error
	// Get returns the entry.
	Get(key string) (std.Option[Entry], error)
}
