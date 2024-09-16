package json

import (
	"encoding/json"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"log/slog"
)

// GetOrZero either returns the unmarshalled value or silently returns the zero value.
func GetOrZero[T any](store blob.Store, key string) T {
	optT, err := Get[T](store, key)
	if err != nil {
		var zero T
		slog.Error("GetOrZero ignored unmarshal error", "err", err)
		return zero
	}

	return optT.Unwrap()
}

// Get reads the value identified by key using the same unmarshalling as JSON [Repository].
func Get[T any](store blob.Store, key string) (std.Option[T], error) {
	var v T
	optBuf, err := blob.Get(store, key)
	if err != nil {
		return std.None[T](), err
	}

	if optBuf.IsNone() {
		return std.None[T](), nil
	}

	if err := json.Unmarshal(optBuf.Unwrap(), &v); err != nil {
		return std.None[T](), err
	}

	return std.Some(v), nil
}

// Put inserts or updates the value identified by key using the same serialization as JSON [Repository].
// Keep in mind, that you should probably never mix a [Repository] of a single type and using this method to write
// other types.
func Put[T any](store blob.Store, key string, value T) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return blob.Put(store, key, buf)
}
