// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/btree"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xslices"
)

const debug = true

type DB struct {
	store     blob.Store
	forward   *btree.BTreeG[fixedBinaryTriple]
	backward  *btree.BTreeG[fixedBinaryTriple]
	strTable  *strTable
	mutex     sync.RWMutex
	resolvers []Resolver
	resources concurrent.RWMap[Namespace, Resources]
}

// NewDB creates a new RBAC database backed by the given blob store.
// It encodes triples as strings using : as separator and uses multiple in-memory b-trees to efficiently lookup triples
// by source or by target. It avoids any further allocations by interning and re-using strings.
func NewDB(store blob.Store) (*DB, error) {
	db := &DB{
		strTable: newStrTable(),
		forward: btree.NewBTreeG[fixedBinaryTriple](func(a, b fixedBinaryTriple) bool {
			return bytes.Compare(a[:], b[:]) < 0
		}),
		backward: btree.NewBTreeG[fixedBinaryTriple](func(a, b fixedBinaryTriple) bool {
			return bytes.Compare(a[:], b[:]) < 0
		}),
		store: store,
	}

	slog.Info("loading ReBAC database from blob store...")
	count := 0
	for key, err := range store.List(context.Background(), blob.ListOptions{}) {
		if err != nil {
			return nil, err
		}

		triple, err := db.decode(key)
		if err != nil {
			return nil, err
		}

		bTriple := db.intoBinary(triple)
		db.forward.Set(bTriple.Binary())
		db.backward.Set(bTriple.reverse().Binary())
		count++
	}

	slog.Info("indexing of ReBAC database completed", "entries", count)

	return db, nil
}

func (db *DB) AddResolver(resolver Resolver) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.resolvers = append(db.resolvers, resolver)
}

// Resolve checks if the database allows the given triple. If the triple is not directly contained within the database,
// it will be checked against all registered resolvers.
// Resolvers are evaluated in registration order: first positive match wins.
func (db *DB) Resolve(triple Triple) (bool, error) {
	if ok, err := db.Contains(triple); ok || err != nil {
		return ok, err
	}

	db.mutex.RLock()
	resolvers := db.resolvers
	db.mutex.RUnlock()

	for _, resolver := range resolvers {
		allowed, err := resolver(db, triple)
		if err != nil {
			return false, err
		}

		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// Contains checks if the database has the given exact triple.
func (db *DB) Contains(tuple Triple) (bool, error) {
	// security note: tryIntoBinary never interns any string to avoid simple DOS attack vectors by just
	// requesting infinite numbers of checks, which otherwise would fill our string table with garbage.
	b, ok := db.tryIntoBinary(tuple)
	if !ok {
		return false, nil
	}

	_, ok = db.forward.Get(b.Binary())
	return ok, nil
}

func optimizerWantsReverse(query Query) bool {
	t := query.triple
	if t.Source.Namespace != "" && t.Source.Instance != "" {
		// forward
		return false
	}

	if t.Target.Namespace != "" && t.Target.Instance != "" {
		//backward
		return true
	}

	if t.Target.Namespace != "" {
		//backward
		return false
	}

	return true
}

// Query retrieves all triples that match the given query. This does not involve any resolving. Leaving
// any namespace, instance or relation empty triggers an according range search. If no source namespace is
// specified, the inverse index is used. At worst, a full table scan is performed if no namespace is specified,
// which is performed internally as a fixed vector comparison and will be faster than conventional string
// Triple equivalence.
func (db *DB) Query(query Query) iter.Seq2[Triple, error] {
	return func(yield func(Triple, error) bool) {
		var tree *btree.BTreeG[fixedBinaryTriple]
		var reverse bool

		if optimizerWantsReverse(query) {
			tree = db.backward
			reverse = true
		} else {
			tree = db.forward
			reverse = false
		}

		if debug {
			slog.Info("db optimizer picked index", "reverse", reverse)
		}

		bTriple, ok := db.tryIntoBinary(query.triple)
		if !ok {
			// some string is not interned; therefore, the result must be the empty set
			return
		}

		queryTriple := bTriple
		if reverse {
			bTriple = bTriple.reverse()
		}

		prefixBytes := prefixSlice(bTriple.Binary())
		emptyPrefix := len(prefixBytes) == 0

		// we may get query triples with [a b nil c d], and the pivot prefix must stop at [a b] or [c d]
		var prefixPivot fixedBinaryTriple
		copy(prefixPivot[:], prefixBytes)

		var debugVisited int
		var debugStart time.Time

		if debug {
			debugStart = time.Now()
		}

		var groupByRelation map[bRelation]struct{}
		if query.groupByRelation {
			groupByRelation = make(map[bRelation]struct{})
		}

		const next = true
		tree.Ascend(prefixPivot, func(item fixedBinaryTriple) bool {
			if debug {
				debugVisited++
			}

			it := item.Unwrap()
			if reverse {
				it = it.reverse()
			}

			// exit early if the prefix does not match anymore
			if !emptyPrefix && !bytes.HasPrefix(item[:], prefixBytes) {
				return false
			}

			// check if we match the query triple, which is not just a prefix but may contain arbitrary query holes
			if queryTriple.A.Namespace != 0 && it.A.Namespace != queryTriple.A.Namespace {
				return next
			}

			if queryTriple.A.Instance != 0 && it.A.Instance != queryTriple.A.Instance {
				return next
			}

			if queryTriple.Relation != 0 && it.Relation != queryTriple.Relation {
				return next
			}

			if queryTriple.B.Namespace != 0 && it.B.Namespace != queryTriple.B.Namespace {
				return next
			}

			if queryTriple.B.Instance != 0 && it.B.Instance != queryTriple.B.Instance {
				return next
			}

			if query.groupByRelation {

			}

			if !query.hasGroupBy() {
				triple := db.fromBinary(it)
				return yield(triple, nil)
			}

			if query.groupByRelation {
				groupByRelation[it.Relation] = struct{}{}
			}

			return true // group by must iterate over all to collect
		})

		if debug {
			slog.Info("rebac db query done", "visited", debugVisited, "duration", time.Since(debugStart).String())
		}

		if query.groupByRelation {
			for relation := range groupByRelation {
				if !yield(db.fromBinary(binaryTriple{
					Relation: relation,
				}), nil) {
					return
				}
			}
		}
	}
}

// Delete removes the exact indexed tuple.
func (db *DB) Delete(tuple Triple) error {

	bTriple, ok := db.tryIntoBinary(tuple)
	if !ok {
		// not indexed
		return nil
	}

	_, has := db.forward.Get(bTriple.Binary())
	// do not issue blind deletes which may accumulate in current persistent storage in WAL
	if !has {
		return nil
	}

	if err := db.store.Delete(context.Background(), db.encode(tuple)); err != nil {
		return err
	}

	db.forward.Delete(bTriple.Binary())
	db.backward.Delete(bTriple.reverse().Binary())

	return nil
}

// DeleteByQuery deletes all tuples that matches the given query. See also [DB.Query] for the query rules.
func (db *DB) DeleteByQuery(query Query) error {
	// our current btree implementation cannot mutate while iterating, so we need to collect all tuples first
	tmp, err := xslices.Collect2(db.Query(query))
	if err != nil {
		return err
	}

	for _, triple := range tmp {
		if err := db.Delete(triple); err != nil {
			return err
		}
	}

	return nil
}

// prefixSlice checks each uint32 in the fixed binary triple and returns the resulting slice until and exclusive
// the first found zero pointer.
func prefixSlice(p fixedBinaryTriple) []byte {
	// 5 fields à 4 Bytes = 20 Bytes total
	for i := 0; i < 5; i++ {
		offset := i * 4
		// check for zero pointer
		if p[offset] == 0 && p[offset+1] == 0 && p[offset+2] == 0 && p[offset+3] == 0 {
			return p[:offset]
		}
	}

	return p[:]
}

// Put just inserts the given tuple in an idempotent fashion.
func (db *DB) Put(tuple Triple) error {
	b := db.intoBinary(tuple)

	// do not issue blind inserts which may accumulate in current persistent storage
	_, has := db.forward.Get(b.Binary())
	if has {
		return nil
	}

	encoded := db.encode(tuple)

	if err := blob.Put(db.store, encoded, nil); err != nil {
		return err
	}

	db.forward.Set(b.Binary())
	db.backward.Set(b.reverse().Binary())

	return nil
}

// PutAll always reflects the persistent state of tuples. If Put returns an error an undefined amount
// of Tuples may have been applied and persisted. All given tuples are applied in a batch which is more
// efficient than applying them one by one.
func (db *DB) PutAll(it iter.Seq2[Triple, error]) error {
	for tuple, err := range it {
		if err != nil {
			return err
		}

		b := db.intoBinary(tuple)
		_, has := db.forward.Get(b.Binary())
		if has {
			continue
		}

		encoded := db.encode(tuple)

		if err := blob.Put(db.store, encoded, nil); err != nil {
			return err
		}

		db.forward.Set(b.Binary())
		db.backward.Set(b.reverse().Binary())
	}

	return nil
}

// Count estimates the number of triples in the database.
func (db *DB) Count() (int64, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	return int64(db.forward.Len()), nil
}

func (db *DB) All() iter.Seq2[Triple, error] {
	return func(yield func(Triple, error) bool) {
		db.forward.Walk(func(items []fixedBinaryTriple) bool {
			for _, item := range items {
				triple := db.fromBinary(item.Unwrap())
				return yield(triple, nil)
			}

			return true
		})
	}
}

// intoBinary interns any given string and encodes the triple into a fixed binary vector representation.
// Empty strings will be encoded as non-interned 0 pointers.
func (db *DB) intoBinary(t Triple) binaryTriple {
	var b binaryTriple

	srcNsID := db.strTable.Intern(string(t.Source.Namespace))
	srcInstID := db.strTable.Intern(string(t.Source.Instance))
	relID := db.strTable.Intern(string(t.Relation))
	tgtNsID := db.strTable.Intern(string(t.Target.Namespace))
	tgtInstID := db.strTable.Intern(string(t.Target.Instance))

	b.A.Namespace = bNamespace(srcNsID)
	b.A.Instance = bInstance(srcInstID)
	b.Relation = bRelation(relID)
	b.B.Namespace = bNamespace(tgtNsID)
	b.B.Instance = bInstance(tgtInstID)

	return b
}

// tryIntoBinary does not mutate the string table to avoid DOS attacks. It returns false if the triple contains
// any unknown string. Empty strings will be encoded as non-interned 0 pointers.
func (db *DB) tryIntoBinary(t Triple) (binaryTriple, bool) {
	var b binaryTriple

	srcNsID, ok := db.strTable.Lookup(string(t.Source.Namespace))
	if !ok {
		return b, false
	}
	srcInstID, ok := db.strTable.Lookup(string(t.Source.Instance))
	if !ok {
		return b, false
	}
	relID, ok := db.strTable.Lookup(string(t.Relation))
	if !ok {
		return b, false
	}
	tgtNsID, ok := db.strTable.Lookup(string(t.Target.Namespace))
	if !ok {
		return b, false
	}
	tgtInstID, ok := db.strTable.Lookup(string(t.Target.Instance))
	if !ok {
		return b, false
	}

	b.A.Namespace = bNamespace(srcNsID)
	b.A.Instance = bInstance(srcInstID)
	b.Relation = bRelation(relID)
	b.B.Namespace = bNamespace(tgtNsID)
	b.B.Instance = bInstance(tgtInstID)

	return b, true
}

func (db *DB) fromBinary(b binaryTriple) Triple {
	return Triple{
		Source: Entity{
			Namespace: Namespace(db.strTable.String(uint32(b.A.Namespace))),
			Instance:  Instance(db.strTable.String(uint32(b.A.Instance))),
		},
		Relation: Relation(db.strTable.String(uint32(b.Relation))),
		Target: Entity{
			Namespace: Namespace(db.strTable.String(uint32(b.B.Namespace))),
			Instance:  Instance(db.strTable.String(uint32(b.B.Instance))),
		},
	}
}

// encode takes the triple and encodes it into a string key using a : as a separator.
// Note that : is allowed in all strings, but the nago framework itself does not use it and we
// expect it to be very uncommon in our identifiers. So there is a fast path, which checks
// if escaping is required, otherwise all strings are just concated together in a pre-allocated string builder.
func (db *DB) encode(tuple Triple) string {
	srcNs := string(tuple.Source.Namespace)
	srcInst := string(tuple.Source.Instance)
	rel := string(tuple.Relation)
	tgtNs := string(tuple.Target.Namespace)
	tgtInst := string(tuple.Target.Instance)

	// fast path: no escaping required
	if strings.IndexByte(srcNs, ':') == -1 &&
		strings.IndexByte(srcInst, ':') == -1 &&
		strings.IndexByte(rel, ':') == -1 &&
		strings.IndexByte(tgtNs, ':') == -1 &&
		strings.IndexByte(tgtInst, ':') == -1 {
		var sb strings.Builder
		sb.Grow(tuple.size())
		sb.WriteString(srcNs)
		sb.WriteByte(':')
		sb.WriteString(srcInst)
		sb.WriteByte(':')
		sb.WriteString(rel)
		sb.WriteByte(':')
		sb.WriteString(tgtNs)
		sb.WriteByte(':')
		sb.WriteString(tgtInst)
		return sb.String()
	}

	// slow path: escape : as ::
	escape := func(s string) string {
		return strings.ReplaceAll(s, ":", "::")
	}

	var sb strings.Builder
	sb.Grow(tuple.size() * 2) // worst case: all colons doubled
	sb.WriteString(escape(srcNs))
	sb.WriteByte(':')
	sb.WriteString(escape(srcInst))
	sb.WriteByte(':')
	sb.WriteString(escape(rel))
	sb.WriteByte(':')
	sb.WriteString(escape(tgtNs))
	sb.WriteByte(':')
	sb.WriteString(escape(tgtInst))
	return sb.String()
}

// decode takes a string key and decodes it back into a Triple.
// It uses a fast path when the key contains exactly 5 parts separated by single colons.
// If the key contains escaped colons (::), it uses a slow path to unescape them.
func (db *DB) decode(key string) (Triple, error) {
	// fast path: try simple split first
	parts := strings.Split(key, ":")
	if len(parts) == 5 {
		return Triple{
			Source: Entity{
				Namespace: Namespace(parts[0]),
				Instance:  Instance(parts[1]),
			},
			Relation: Relation(parts[2]),
			Target: Entity{
				Namespace: Namespace(parts[3]),
				Instance:  Instance(parts[4]),
			},
		}, nil
	}

	// slow path: handle escaped colons (::)
	// when we split "a::b:c" by ":", we get ["a", "", "b", "c"]
	// we need to merge adjacent empty strings back to ":"
	var fields []string
	var current strings.Builder

	for i := 0; i < len(parts); i++ {
		if parts[i] == "" && i+1 < len(parts) && parts[i+1] == "" {
			// found ::, which means an escaped :
			current.WriteByte(':')
			i++ // skip the next empty part
		} else if parts[i] == "" && current.Len() > 0 {
			// single :, this is a field separator
			fields = append(fields, current.String())
			current.Reset()
		} else if parts[i] == "" {
			// single : at field boundary
			fields = append(fields, current.String())
			current.Reset()
		} else {
			current.WriteString(parts[i])
		}
	}
	// don't forget the last field
	fields = append(fields, current.String())

	if len(fields) != 5 {
		return Triple{}, fmt.Errorf("invalid triple key: expected 5 fields, got %d", len(fields))
	}

	return Triple{
		Source: Entity{
			Namespace: Namespace(fields[0]),
			Instance:  Instance(fields[1]),
		},
		Relation: Relation(fields[2]),
		Target: Entity{
			Namespace: Namespace(fields[3]),
			Instance:  Instance(fields[4]),
		},
	}, nil
}

func (db *DB) RegisterResources(res Resources) bool {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.resources.Get(res.Identity()); ok {
		return false
	}

	if res.Identity() == "" {
		return false
	}

	db.resources.Put(res.Identity(), res)
	return true
}

func (db *DB) UnregisterResources(res Namespace) {
	db.resources.Delete(res)
}

func (db *DB) LookupResources(res Namespace) (Resources, bool) {
	return db.resources.Get(res)
}

func (db *DB) AllResources() iter.Seq[Resources] {
	tmp := slices.Collect(db.resources.Values())
	slices.SortFunc(tmp, func(a, b Resources) int {
		return strings.Compare(string(a.Identity()), string(b.Identity()))
	})

	return func(yield func(Resources) bool) {
		for _, r := range tmp {
			if !yield(r) {
				return
			}
		}
	}
}
