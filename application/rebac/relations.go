// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"encoding/binary"
	"iter"
	"slices"
)

// RelationSet is a snapshot of all relations, which a source has on a target, see [DB.Relations]. The zero value
// contains no relation.
type RelationSet struct {
	db  *DB
	ids []uint32 // sorted interned ids
}

// Has decides like [DB.Resolve] for the source, the target and the given relation, at the time the snapshot was
// taken.
func (s RelationSet) Has(relation Relation) bool {
	if s.db == nil || relation == "" {
		return false
	}

	id, ok := s.db.strTable.Lookup(string(relation))
	if !ok {
		return false
	}

	_, found := slices.BinarySearch(s.ids, id)
	return found
}

// Len returns the number of relations in the snapshot.
func (s RelationSet) Len() int {
	return len(s.ids)
}

// All yields the relations of the snapshot in an unspecified but stable order.
func (s RelationSet) All() iter.Seq[Relation] {
	return func(yield func(Relation) bool) {
		if s.db == nil {
			return
		}

		for _, id := range s.ids {
			if !yield(Relation(s.db.strTable.String(id))) {
				return
			}
		}
	}
}

// Relations enumerates all relations, which the source has on the target, exactly as [DB.Resolve] would decide
// them one by one: stored directly, stored for the wildcard target, or derived by a rule from a membership.
// A single pass over the indices is much cheaper than resolving many relations individually, e.g. to decide which
// of a lot of permission dependent views a user may see. An incomplete source or target has no relations.
func (db *DB) Relations(source Entity, target Entity) RelationSet {
	p, ok := db.newProbe(source, "", target)
	if !ok {
		return RelationSet{db: db}
	}

	ids := db.appendRelations(make([]uint32, 0, 128), p, 0)
	slices.Sort(ids)
	return RelationSet{db: db, ids: slices.Compact(ids)}
}

// appendRelations appends the relations of the probe, directly and through the rules, see also [DB.check].
func (db *DB) appendRelations(dst []uint32, p probe, depth int) []uint32 {
	dst = db.appendStoredRelations(dst, p)

	if depth >= MaxRuleDepth {
		return dst
	}

	for _, rule := range *db.rules.Load() {
		if rule.src != p.src.Namespace {
			continue
		}

		var buf [16]bInstance
		for _, inst := range db.appendMemberships(buf[:0], p.src, rule) {
			dst = db.appendRelations(dst, p.via(rule.via, inst), depth+1)
		}
	}

	return dst
}

// appendStoredRelations appends every relation, for which a triple from the source to the target or to the
// wildcard target is stored. The forward index groups the triples of a source by relation and sorts each group by
// target, thus a single triple per relation is visited: if it is not one of the two targets, both are probed by
// seeking. A group with more triples, e.g. the thousands of members of a role, is skipped by seeking to the next
// relation.
func (db *DB) appendStoredRelations(dst []uint32, p probe) []uint32 {
	exact := binaryTriple{A: p.src, B: p.tgt}.Binary()
	wild := p.wildcard()
	wildKey := binaryTriple{A: wild.src, B: wild.tgt}.Binary()
	hasWild := p.hasWildcard()
	prefix := [8]byte(exact[:8])
	exactTarget := [8]byte(exact[12:20])
	wildTarget := [8]byte(wildKey[12:20])

	it := db.forward.Iter()
	defer it.Release()

	ok := it.Seek(exact)
	for ok {
		item := it.Item()
		if [8]byte(item[:8]) != prefix {
			return dst
		}

		rel := binary.LittleEndian.Uint32(item[8:12])
		target := [8]byte(item[12:20])

		if target == exactTarget || (hasWild && target == wildTarget) {
			dst = append(dst, rel)

			// usually the next triple already belongs to the next relation
			ok = it.Next()
			if following := it.Item(); !ok || [12]byte(following[:12]) != [12]byte(item[:12]) {
				continue
			}
		} else {
			// the group is sorted by target, thus probe both candidates directly
			var key fixedBinaryTriple
			copy(key[:12], item[:12])
			copy(key[12:], exactTarget[:])
			if ok = it.Seek(key); ok && it.Item() == key {
				dst = append(dst, rel)
			} else if hasWild {
				copy(key[12:], wildTarget[:])
				if ok = it.Seek(key); ok && it.Item() == key {
					dst = append(dst, rel)
				}
			}
		}

		// the probes may have moved anywhere behind the group: continue at the next relation, which also skips
		// the remaining triples of the group, e.g. the thousands of members of a role
		next, hasNext := item.afterRelation()
		if !hasNext {
			return dst
		}

		ok = it.Seek(next)
	}

	return dst
}
