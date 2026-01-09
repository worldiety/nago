// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"strings"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
)

// Indexer derives information from a domain event and stores it into a separate composite index, where the
// entry is stored as <primary>-<seqid> tuples without any associated values. Note that the index and event
// stores are logically separated and may get corrupted, e.g. if the event has been stored but cannot get
// indexed.
type Indexer[Evt any] interface {
	// Insert is called after the event has been stored and processed and can be inserted into the index.
	// This is usually an O(log(n)) effort.
	Insert(envelope Envelope[Evt]) error

	// Remove removes the given event from the index. Usually this only involves an O(log(n)) lookup
	// for this specific entry. If multiple Indexer are defined each of them is called and must remove
	// there entry as well.
	Remove(envelope Envelope[Evt]) error

	// GroupByPrimaryAsString returns all primary keys as a sorted string sequence.
	GroupByPrimaryAsString() (iter.Seq2[string, int], error)

	GroupByPrimary(primary string) iter.Seq2[SeqKey, error]

	Info() IndexerInfo
}

type IdxID string
type IndexerInfo struct {
	ID IdxID
	// Name of the indexer or field to index
	Name        string
	Description string
}

type idxKey[Primary ~string] string

func newIdxKey[Primary ~string](p Primary, seqId SeqKey) idxKey[Primary] {
	return idxKey[Primary](string(p) + "-" + string(seqId))
}

func (k idxKey[Primary]) Parse() (Primary, SeqKey, error) {
	pos := strings.LastIndex(string(k), "-")
	if pos < 1 {
		return "", "", fmt.Errorf("invalid key: %s", k)
	}

	return Primary(k[:pos]), SeqKey(k[pos+1:]), nil
}

type StoreIndex[Primary ~string, Evt any] struct {
	*data.CompositeIndex[Primary, SeqKey]
	reader func(Envelope[Evt]) (Primary, error)
	info   IndexerInfo
}

type PrimaryReader[Primary ~string, Evt any] func(Envelope[Evt]) (Primary, error)

// NewStoreIndex returns a new implementation for a store indexer based on a simple primary reader function.
// If the returned error is [fs.SkipAll], the event is just omitted and no error is returned.
func NewStoreIndex[Primary ~string, Evt any](idxStore blob.Store, reader PrimaryReader[Primary, Evt]) *StoreIndex[Primary, Evt] {
	idx := data.NewCompositeIndex[Primary, SeqKey](idxStore)
	idx.SetKeyDecoder(func(s string) (data.CompositeKey[Primary, SeqKey], error) {
		pk, seqId, err := idxKey[Primary](s).Parse()
		return data.CompositeKey[Primary, SeqKey]{Primary: pk, Secondary: seqId}, err
	})

	idx.SetKeyEncoder(func(key data.CompositeKey[Primary, SeqKey]) (string, error) {
		s := newIdxKey(key.Primary, key.Secondary)
		return string(s), nil
	})

	return &StoreIndex[Primary, Evt]{
		CompositeIndex: idx,
		reader:         reader,
	}
}

func (idx *StoreIndex[Primary, Evt]) Insert(envelope Envelope[Evt]) error {
	pk, err := idx.reader(envelope)
	if err != nil {
		if errors.Is(err, fs.SkipAll) {
			// omit by definition
			return nil
		}

		return err
	}

	return idx.Put(pk, envelope.Key)
}

func (idx *StoreIndex[Primary, Evt]) Remove(envelope Envelope[Evt]) error {
	pk, err := idx.reader(envelope)
	if err != nil {
		if errors.Is(err, fs.SkipAll) {
			// omit by definition
			return nil
		}

		return err
	}

	return idx.Delete(context.Background(), pk, envelope.Key)
}

func (idx *StoreIndex[Primary, Evt]) GroupByPrimaryAsString() (iter.Seq2[string, int], error) {
	it, err := idx.CompositeIndex.GroupPrimary(context.Background())
	if err != nil {
		return nil, err
	}

	return func(yield func(string, int) bool) {
		for p, count := range it {
			if !yield(string(p), count) {
				return
			}
		}
	}, nil
}

func (idx *StoreIndex[Primary, Evt]) SetInfo(info IndexerInfo) {
	idx.info = info
}

func (idx *StoreIndex[Primary, Evt]) Info() IndexerInfo {
	return idx.info
}

func (idx *StoreIndex[Primary, Evt]) GroupByPrimary(primary string) iter.Seq2[SeqKey, error] {
	return func(yield func(SeqKey, error) bool) {
		for key, err := range idx.CompositeIndex.AllByPrimary(context.Background(), Primary(primary)) {
			if !yield(key.Secondary, err) {
				return
			}
		}
	}
}
