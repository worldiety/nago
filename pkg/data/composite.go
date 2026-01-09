// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package data

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"strings"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xmaps"
)

type CompositeKey[A, B ~string] struct {
	Primary   A
	Secondary B
}

func (c CompositeKey[A, B]) String() string {
	return string(c.Primary) + "." + string(c.Secondary)
}

// CompositeIndex wraps a store based on the given [blob.Store] and just stores keys.
// It does not store any payload within the store itself and is optimized to manage only pairs of Primary/Secondary
// tuples. See also [NewComposite] and [NewCompositeIndex].
type CompositeIndex[A, B ~string] struct {
	store      blob.Store
	keyEncoder func(key CompositeKey[A, B]) (string, error)
	keyDecoder func(string) (CompositeKey[A, B], error)
}

// NewComposite creates or opens a new index for the given repositories in a type safe way.
// If the given repositories are of type [NotifyRepository] an automatic synchronous cascade-delete will be performed.
func NewComposite[AEntity Aggregate[AID], AID ~string, BEntity Aggregate[BID], BID ~string](stores blob.Stores, repoA Repository[AEntity, AID], repoB Repository[BEntity, BID]) (*CompositeIndex[AID, BID], error) {
	idxStore, err := stores.Open("index."+repoA.Name()+"."+repoB.Name(), blob.OpenStoreOptions{
		Type: blob.EntityStore,
	})
	if err != nil {
		return nil, err
	}

	idx := NewCompositeIndex[AID, BID](idxStore)

	if evtRepo, ok := repoA.(NotifyRepository[AEntity, AID]); ok {
		evtRepo.AddDeletedObserver(func(repository Repository[AEntity, AID], deleted Deleted[AID]) error {
			return idx.DeleteAllPrimary(context.Background(), deleted.ID)
		})
	} else {
		slog.Info("composite index cannot automatically delete cascade from primary repository: not a NotifyRepository", "type", fmt.Sprintf("%T", repoA))
	}

	if evtRepo, ok := repoB.(NotifyRepository[BEntity, BID]); ok {
		evtRepo.AddDeletedObserver(func(repository Repository[BEntity, BID], deleted Deleted[BID]) error {
			return idx.DeleteAllSecondary(context.Background(), deleted.ID)
		})
	} else {
		slog.Info("composite index cannot automatically delete cascade from secondary repository: not a NotifyRepository", "type", fmt.Sprintf("%T", repoB))
	}

	return idx, nil
}

// NewCompositeIndex creates a new index. You probably want [NewComposite].
func NewCompositeIndex[A, B ~string](store blob.Store) *CompositeIndex[A, B] {
	return &CompositeIndex[A, B]{
		store: store,
	}
}

// SetKeyEncoder replaces the default encoding behavior. Note that it is valid to use an empty B
// e.g. to encode and express a prefix search. It is important that the encoder still appends the correct
// separator char (if any).
func (idx *CompositeIndex[A, B]) SetKeyEncoder(encoder func(key CompositeKey[A, B]) (string, error)) {
	idx.keyEncoder = encoder
}

func (idx *CompositeIndex[A, B]) SetKeyDecoder(decoder func(string) (CompositeKey[A, B], error)) {
	idx.keyDecoder = decoder
}

func (idx *CompositeIndex[A, B]) encode(a A, b B) (string, error) {
	if idx.keyEncoder != nil {
		return idx.keyEncoder(CompositeKey[A, B]{a, b})
	}

	// fallback to the hardcoded default behavior
	if strings.Contains(string(a), ".") {
		return "", fmt.Errorf("invalid key composite a: must not contain '.': %s", a)
	}

	if strings.Contains(string(b), ".") {
		return "", fmt.Errorf("invalid key composite b: must not contain '.': %s", b)
	}

	return string(a) + "." + string(b), nil
}

func (idx *CompositeIndex[A, B]) decode(key string) (CompositeKey[A, B], error) {
	if idx.keyDecoder != nil {
		return idx.keyDecoder(key)
	}

	// default fallback behavior
	tokens := strings.Split(key, ".")
	if len(tokens) != 2 {
		return CompositeKey[A, B]{}, fmt.Errorf("invalid tokens: %s", key)
	}
	return CompositeKey[A, B]{A(tokens[0]), B(tokens[1])}, nil
}

func (idx *CompositeIndex[A, B]) Put(a A, b B) error {
	key, err := idx.encode(a, b)
	if err != nil {
		return err
	}

	if err := blob.Put(idx.store, key, nil); err != nil {
		return fmt.Errorf("cannot put composite index into store: %w", err)
	}

	return nil
}

func (idx *CompositeIndex[A, B]) Delete(ctx context.Context, a A, b B) error {
	key, err := idx.encode(a, b)
	if err != nil {
		return err
	}

	return idx.store.Delete(ctx, key)
}

func (idx *CompositeIndex[A, B]) Exists(ctx context.Context, a A, b B) (bool, error) {
	key, err := idx.encode(a, b)
	if err != nil {
		return false, err
	}

	return idx.store.Exists(ctx, key)
}

// DeleteAllPrimary removes all those entries which start with a.
func (idx *CompositeIndex[A, B]) DeleteAllPrimary(ctx context.Context, a A) error {
	for key, err := range idx.AllByPrimary(ctx, a) {
		if err != nil {
			return err
		}
		key, err := idx.encode(key.Primary, key.Secondary)
		if err != nil {
			return err
		}

		if err := idx.store.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

// DeleteAllSecondary requires an O(n) walk over the entire index to remove all secondary entries.
// Depending on how large the index is, this may be acceptable. The default store implementations keep
// the keys in memory anyway, thus you just have a lot of pointers passed around in practice.
func (idx *CompositeIndex[A, B]) DeleteAllSecondary(ctx context.Context, b B) error {
	for key, err := range idx.All(ctx) {
		if err != nil {
			return err
		}

		if key.Secondary == b {
			strKey, err := idx.encode(key.Primary, key.Secondary)
			if err != nil {
				return err
			}

			if err := idx.store.Delete(ctx, strKey); err != nil {
				return err
			}
		}
	}

	return nil
}

// Count just returns the total amount of entries in the index.
func (idx *CompositeIndex[A, B]) Count(ctx context.Context) (int64, error) {
	return blob.Count(ctx, idx.store)
}

// CountByPrimary just returns the total amount of primary entries in the index.
func (idx *CompositeIndex[A, B]) CountByPrimary(ctx context.Context, primary A) (int64, error) {
	var count int64
	for _, err := range idx.AllByPrimary(ctx, primary) {
		if err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
}

// All just loops of the entire key set.
func (idx *CompositeIndex[A, B]) All(ctx context.Context) iter.Seq2[CompositeKey[A, B], error] {
	return idx.AllByPrefix(ctx, "")
}

// Clear just removes all entries.
func (idx *CompositeIndex[A, B]) Clear(ctx context.Context) error {
	return blob.DeleteAll(idx.store)
}

// AllByPrefix can be a partial a.b* and is the fastest reduction of effort you can get from the underlying store.
// See also AllByPrimary for a bit more safety.
func (idx *CompositeIndex[A, B]) AllByPrefix(ctx context.Context, prefix string) iter.Seq2[CompositeKey[A, B], error] {
	return func(yield func(CompositeKey[A, B], error) bool) {
		for s, err := range idx.store.List(ctx, blob.ListOptions{
			Prefix: prefix,
		}) {
			if err != nil {
				if !yield(CompositeKey[A, B]{}, err) {
					return
				}

				continue
			}

			key, err := idx.decode(s)
			if err != nil {
				if !yield(CompositeKey[A, B]{}, err) {
					return
				}

				continue
			}

			if !yield(key, nil) {
				return
			}
		}
	}
}

// AllByPrimary safely returns only children of A.
func (idx *CompositeIndex[A, B]) AllByPrimary(ctx context.Context, prefix A) iter.Seq2[CompositeKey[A, B], error] {
	// security note: ensure that we actually will find only by a and not some a' which also starts with a via injection
	sKey, err := idx.encode(prefix, "")
	if err != nil {
		return xiter.WithError[CompositeKey[A, B]](err)
	}

	return idx.AllByPrefix(ctx, sKey)
}

// AllBySecondary requires an O(n) walk over the entire index to find all keys which have B as secondary. See also
// [CompositeIndex.DeleteAllSecondary].
func (idx *CompositeIndex[A, B]) AllBySecondary(ctx context.Context, secondary B) iter.Seq2[CompositeKey[A, B], error] {
	return func(yield func(CompositeKey[A, B], error) bool) {
		for key, err := range idx.All(ctx) {
			if err != nil {
				if !yield(key, err) {
					return
				}

				continue
			}

			if key.Secondary == secondary {
				if !yield(key, nil) {
					return
				}
			}
		}
	}
}

// GroupPrimary returns a unique and sorted set of all primary keys and the number of contained associations.
func (idx *CompositeIndex[A, B]) GroupPrimary(ctx context.Context) (iter.Seq2[A, int], error) {
	tmp := make(map[A]int)
	for key, err := range idx.All(ctx) {
		if err != nil {
			return nil, err
		}

		tmp[key.Primary] = tmp[key.Primary] + 1
	}

	return func(yield func(A, int) bool) {
		for _, a := range xmaps.SortedKeys(tmp) {
			if !yield(a, tmp[a]) {
				return
			}
		}
	}, nil
}
