// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package json

import (
	"encoding/json"
	"log/slog"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/blob"
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
func Get[T any](store blob.Reader, key string) (option.Opt[T], error) {
	var v T
	optBuf, err := blob.Get(store, key)
	if err != nil {
		return option.None[T](), err
	}

	if optBuf.IsNone() {
		return option.None[T](), nil
	}

	if err := json.Unmarshal(optBuf.Unwrap(), &v); err != nil {
		return option.None[T](), err
	}

	return option.Some(v), nil
}

// Put inserts or updates the value identified by key using the same serialization as JSON [Repository].
// Keep in mind, that you should probably never mix a [Repository] of a single type and using this method to write
// other types.
func Put[T any](store blob.Writer, key string, value T) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return blob.Put(store, key, buf)
}
