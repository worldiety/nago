// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"encoding/binary"
	"math"
	"slices"
)

type memberResolver struct {
	src    Namespace
	member Namespace
}

// GlobalGrants is a snapshot of all relations, which a source has on the global wildcard target, see
// [DB.GlobalGrants]. The zero value contains no relation.
type GlobalGrants struct {
	db        *DB
	relations []uint32 // sorted interned ids
}

// Has decides like [DB.Resolve] for the triple of the snapshot source, the given relation and the global wildcard
// target, at the time the snapshot was taken.
func (g GlobalGrants) Has(relation Relation) bool {
	if g.db == nil || relation == "" {
		return false
	}

	id, ok := g.db.strTable.Lookup(string(relation))
	if !ok {
		return false
	}

	_, found := slices.BinarySearch(g.relations, id)
	return found
}

// GlobalGrants collects all relations, which the source has on the global wildcard target, either directly or
// through the resolvers added by [DB.AddMemberResolver]. This is much cheaper than resolving many relations one by
// one, e.g. to decide which of a lot of permission dependent views a user can see. It returns false, if the
// decision of [DB.Resolve] cannot be enumerated, because a resolver has been added by [DB.AddResolver] or the
// source is incomplete. In that case, each relation must be resolved individually.
func (db *DB) GlobalGrants(source Entity) (GlobalGrants, bool) {
	if source.Namespace == "" || source.Instance == "" {
		return GlobalGrants{}, false
	}

	members := make([]Namespace, 0, 4)
	db.mutex.RLock()
	enumerable := len(db.resolvers) == len(db.memberResolvers)
	for _, resolver := range db.memberResolvers {
		if resolver.src == source.Namespace {
			members = append(members, resolver.member)
		}
	}
	db.mutex.RUnlock()

	if !enumerable {
		return GlobalGrants{}, false
	}

	memberNamespaces := make([]uint32, 0, 4)
	t := db.strTable
	t.mutex.RLock()
	srcNs, ok1 := t.lookupLocked(string(source.Namespace))
	srcInst, ok2 := t.lookupLocked(string(source.Instance))
	globalNs, ok3 := t.lookupLocked(string(Global))
	allInstances, ok4 := t.lookupLocked(string(AllInstances))
	member, hasMember := t.lookupLocked(string(Member))
	for _, ns := range members {
		// an unknown namespace cannot have any members
		if id, ok := t.lookupLocked(string(ns)); ok {
			memberNamespaces = append(memberNamespaces, id)
		}
	}
	t.mutex.RUnlock()

	g := GlobalGrants{db: db}
	if !(ok1 && ok2 && ok3 && ok4) {
		// a string which has never been interned cannot be part of any grant
		return g, true
	}

	var target [8]byte
	binary.LittleEndian.PutUint32(target[0:4], globalNs)
	binary.LittleEndian.PutUint32(target[4:8], allInstances)

	g.relations = db.appendGlobalRelations(g.relations, srcNs, srcInst, member, hasMember, target)

	if hasMember {
		// the backward index stores target, relation and source, thus all memberships of the source are a single
		// range per member namespace, see also containsViaMember
		for _, memberNs := range memberNamespaces {
			var pivot fixedBinaryTriple
			binary.LittleEndian.PutUint32(pivot[0:4], srcNs)
			binary.LittleEndian.PutUint32(pivot[4:8], srcInst)
			binary.LittleEndian.PutUint32(pivot[8:12], member)
			binary.LittleEndian.PutUint32(pivot[12:16], memberNs)

			db.backward.Ascend(pivot, func(item fixedBinaryTriple) bool {
				if [16]byte(item[:16]) != [16]byte(pivot[:16]) {
					return false
				}

				memberInst := binary.LittleEndian.Uint32(item[16:20])
				g.relations = db.appendGlobalRelations(g.relations, memberNs, memberInst, member, true, target)
				return true
			})
		}
	}

	slices.Sort(g.relations)
	return g, true
}

// appendGlobalRelations appends the relations of all triples from the given source to the target.
func (db *DB) appendGlobalRelations(dst []uint32, ns uint32, inst uint32, member uint32, hasMember bool, target [8]byte) []uint32 {
	var pivot fixedBinaryTriple
	binary.LittleEndian.PutUint32(pivot[0:4], ns)
	binary.LittleEndian.PutUint32(pivot[4:8], inst)
	prefix := [8]byte(pivot[:8])

	for {
		skipped := false
		db.forward.Ascend(pivot, func(item fixedBinaryTriple) bool {
			if [8]byte(item[:8]) != prefix {
				return false
			}

			relation := binary.LittleEndian.Uint32(item[8:12])
			if hasMember && relation == member {
				// A role may have thousands of members, which are never relevant here. The btree is ordered by the
				// big endian interpretation of each word, thus the next relation starts at the successor.
				word := binary.BigEndian.Uint32(item[8:12])
				if word == math.MaxUint32 {
					return false
				}

				pivot = fixedBinaryTriple{}
				copy(pivot[:8], prefix[:])
				binary.BigEndian.PutUint32(pivot[8:12], word+1)
				skipped = true
				return false
			}

			if [8]byte(item[12:20]) == target {
				dst = append(dst, relation)
			}

			return true
		})

		if !skipped {
			return dst
		}
	}
}
