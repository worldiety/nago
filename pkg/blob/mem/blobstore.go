package mem

import (
	"bytes"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"sync"
)

type Entries = map[string][]byte

// BlobStore provides an in-memory implementation without transactions.
// The transactions are just fake implementations to satisfy the contract and respect the read/write property.
// However, the store itself is at least thread safe.
type BlobStore struct {
	values map[string][]byte
	mutex  sync.RWMutex
}

// NewBlobStore creates a new in-memory store.
func NewBlobStore() *BlobStore {
	return &BlobStore{values: map[string][]byte{}}
}

func From(values map[string][]byte) *BlobStore {
	return &BlobStore{values: values}
}

func (b *BlobStore) Update(f func(blob.Tx) error) error {
	tx := &memTx{parent: b}
	if err := f(tx); err != nil {
		return err
	}

	return nil
}

func (b *BlobStore) View(f func(blob.Tx) error) error {
	tx := &memTx{parent: b}
	if err := f(tx); err != nil {
		return err
	}

	return nil
}

type memTx struct {
	parent   *BlobStore
	readOnly bool
}

func (m *memTx) Each(yield func(blob.Entry, error) bool) {
	m.parent.mutex.RLock()
	defer m.parent.mutex.RUnlock()

	for key, val := range m.parent.values {
		if !yield(blob.Entry{
			Key: key,
			Open: func() (io.ReadCloser, error) {
				return readerCloser{Reader: bytes.NewReader(val)}, nil
			},
		}, nil) {
			return
		}
	}
}

func (m *memTx) Delete(key string) error {
	m.parent.mutex.Lock()
	defer m.parent.mutex.Unlock()

	if m.readOnly {
		return fmt.Errorf("transaction is read only")
	}

	delete(m.parent.values, key)
	return nil
}

func (m *memTx) Put(entry blob.Entry) error {
	m.parent.mutex.Lock()
	defer m.parent.mutex.Unlock()

	if m.readOnly {
		return fmt.Errorf("transaction is read only")
	}

	r, err := entry.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	m.parent.values[entry.Key] = buf

	return nil
}

func (m *memTx) Get(key string) (std.Option[blob.Entry], error) {
	m.parent.mutex.RLock()
	defer m.parent.mutex.RUnlock()

	buf, ok := m.parent.values[key]
	if !ok {
		return std.None[blob.Entry](), nil
	}

	return std.Some(blob.Entry{
		Key: key,
		Open: func() (io.ReadCloser, error) {
			return readerCloser{Reader: bytes.NewReader(buf)}, nil
		},
	}), nil
}
