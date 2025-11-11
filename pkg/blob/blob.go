// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package blob

import (
	"bytes"
	"context"
	"io"
	"iter"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
)

type ListOptions struct {
	// If non-zero, the result set will only contain keys which starts with the given prefix.
	Prefix string
	// If non-zero, MinInc marks the inclusive minimal starting key in the result set.
	MinInc string
	// If non-zero, MaxInc marks the inclusive minimal ending key in the result set.
	MaxInc string

	// If not zero, returns at most the amount of given entries.
	Limit int
}

type Reader interface {
	// NewReader opens the blob to be read.
	NewReader(ctx context.Context, key string) (option.Opt[io.ReadCloser], error)
}

type Writer interface {
	// NewWriter open the blob to be created or overwritten. Either of them will only happen
	// if the writer has been closed and the context has not been cancelled.
	// A Write is always atomic and implementations must ensure, that
	// a partial write is never visible.
	NewWriter(ctx context.Context, key string) (io.WriteCloser, error)
}

type ObjectWriter interface {
	Put(ctx context.Context, key string, r io.Reader) (int64, error)
}

type ObjectReader interface {
	Get(ctx context.Context, key string, r io.Writer) (int64, error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Deleter interface {
	// Delete removes the denoted entry. It is not an error to remove a non-existent file.
	Delete(ctx context.Context, key string) error
}

type Info struct {
	Size    int64     // -1 if Size is unavailable.
	ModTime time.Time // zero value if unavailable
	Sys     any       // contains any implementation specific details about the object.
}

type StatReader interface {
	// Stat reads some metadata for the given key.
	Stat(ctx context.Context, key string) (option.Opt[Info], error)
}

type Counter interface {
	Count(ctx context.Context) (int64, error)
}

// Store represents a single bucket store for blobs. Note, that individual methods are thread safe, however
// it is not possible to represent transactions.
// This limitation is intentionally, because neither simple implementations (an ordinary filesystem) nor
// scaling out implementations (eventual consistent clustered cloud storage) support proper transactions.
// Providing a transactional closure will also either provide surprising behavior or deadlocks by definition
// (start a read transaction and nest a write transaction - what shall happen?). Massive scalable cloud systems
// are also only scalable and fast, if used in a non-transactional and eventual-consistent way.
type Store interface {
	// List takes a snapshot of all available entries and returns an iterator for it.
	// While iterating, any operation on the dataset can be performed without blocking, however
	// these changes must not cause the iterator to return garbage (like missed or doubled entries).
	// Note that this may become very inefficient, when used on very large datasets containing
	// millions or even billions of entries. The order of the returned keys is sorted lexicographically from
	// smallest to largest. Thus, the smallest key in a Store can be efficiently queries, using just a Limit of 1.
	// Implementations must support Prefix and Range filters.
	List(ctx context.Context, opts ListOptions) iter.Seq2[string, error]

	// Exists returns only true, if at least at some time such blob existed. Note, that in concurrent situations
	// such a statement is not very useful.
	Exists(ctx context.Context, key string) (bool, error)

	// Name returns the distinct name. Stores with the same name are considered equal.
	Name() string

	Deleter

	ReadWriter
}

// Read transfers from the store all bytes into the given writer, e.g. into a http response.
func Read(store Reader, key string, dst io.Writer) (exists bool, err error) {
	optR, err := store.NewReader(context.Background(), key)
	if err != nil {
		return false, err
	}

	if optR.IsNone() {
		return false, nil
	}

	r := optR.Unwrap()

	_, err = io.Copy(dst, r)
	return true, err
}

// Write transfers all bytes from the given source into the store, e.g. from a request body.
func Write(store Writer, key string, src io.Reader) (written int64, err error) {
	if objWriter, ok := store.(ObjectWriter); ok {
		return objWriter.Put(context.Background(), key, src)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w, err := store.NewWriter(ctx, key)
	if err != nil {
		return 0, err
	}

	defer std.Try(w.Close, &err)

	n, err := io.Copy(w, src)
	if err != nil {
		cancel() //cancel earlier to mark written data so far as stale
	}

	return n, err
}

// Put is a shorthand function to write small values using a slice into the store. Do not use for large blobs.
func Put(store Writer, key string, value []byte) (err error) {
	_, err = Write(store, key, bytes.NewReader(value))
	return err
}

func Delete(store Store, key string) error {
	return store.Delete(context.Background(), key)
}

// DeleteAll removes all known entries at certain point of time. Due to concurrency and eventual consistency
// the store may not be empty after iteration.
func DeleteAll(store Store) error {
	for key, err := range store.List(context.Background(), ListOptions{}) {
		if err != nil {
			return err
		}

		if err := store.Delete(context.Background(), key); err != nil {
			return err
		}
	}

	return nil
}

// Get is a shortcut function to read small slices from the store. Do not use for large blobs, because it allocates
// the entire blob size without other limits.
func Get(store Reader, key string) (std.Option[[]byte], error) {
	optReader, err := store.NewReader(context.Background(), key)
	if err != nil {
		return std.None[[]byte](), err
	}

	if optReader.IsNone() {
		return std.None[[]byte](), nil
	}

	r := optReader.Unwrap()
	defer r.Close()

	buf, err := io.ReadAll(r)
	if err != nil {
		return std.None[[]byte](), err
	}

	return std.Some(buf), nil
}

func Keys(store Store) ([]string, error) {
	return xslices.Collect2(store.List(context.Background(), ListOptions{}))
}

// Count may enumerate through all entries which may become expensive, and it is just an indicator under concurrency.
// If the given store implements a [Counter] interface, it is used to determine the amount of entries without
// iterating all elements.
func Count(ctx context.Context, store Store) (int64, error) {
	if counter, ok := store.(Counter); ok {
		return counter.Count(ctx)
	}

	var count int64
	for _, err := range store.List(ctx, ListOptions{}) {
		if err != nil {
			return 0, err
		}

		count++
	}

	return count, nil
}
