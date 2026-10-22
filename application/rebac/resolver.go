// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import "encoding/binary"

// A Resolver takes a triple and returns true if the triple is allowed, false otherwise.
// A Resolver must never mutate the database.
type Resolver func(db *DB, triple Triple) (bool, error)

// NewSourceMemberResolver creates a resolver that checks permissions through membership relationships.
//
// The resolution follows these steps:
//  1. Skip if triple.Source.Namespace != srcNamespace (not applicable)
//  2. Find all memberNamespace entities where triple.Source is a Member
//  3. For each found membership, check if that entity grants the requested permission
//  4. Fall back to wildcard check (Target.Instance = "*") if no direct match
//
// Example: User "torben" wants "read" on Document "doc1"
//   - Finds: Role "admin" has Member "torben"
//   - Checks: Role "admin" has "read" on Document "doc1"
//   - Or: Role "admin" has "read" on Document "*"
//
// Note: This resolver is non-recursive (single level).
//
// Parameters:
//   - srcNamespace: The namespace of the source entity (e.g., "user")
//   - memberNamespace: The namespace containing membership relations (e.g., "role")
func NewSourceMemberResolver(srcNamespace Namespace, memberNamespace Namespace) Resolver {
	return func(db *DB, triple Triple) (bool, error) {
		// check if triple source is a user
		if triple.Source.Namespace != srcNamespace {
			return false, nil
		}

		if ok, applicable := db.containsViaMember(memberNamespace, triple); applicable {
			return ok, nil
		}

		return containsViaMemberQuery(db, memberNamespace, triple)
	}
}

// containsViaMember is the fast path of [NewSourceMemberResolver]. The resolver is evaluated by nearly every
// permission check, mostly with a negative result, thus it works directly on the interned and binary encoded
// indices instead of using [DB.Query] and [DB.Contains]. It neither allocates nor converts interned strings back.
// If any of the involved strings is empty, the query semantics of an empty string (match anything) apply, which
// the fast path does not implement, and applicable is false.
func (db *DB) containsViaMember(memberNamespace Namespace, triple Triple) (ok bool, applicable bool) {
	if triple.Source.Namespace == "" || triple.Source.Instance == "" || triple.Relation == "" || triple.Target.Namespace == "" || triple.Target.Instance == "" {
		return false, false
	}

	t := db.strTable
	t.mutex.RLock()
	srcNs, ok1 := t.lookupLocked(string(triple.Source.Namespace))
	srcInst, ok2 := t.lookupLocked(string(triple.Source.Instance))
	memberNs, ok3 := t.lookupLocked(string(memberNamespace))
	member, ok4 := t.lookupLocked(string(Member))
	rel, ok5 := t.lookupLocked(string(triple.Relation))
	tgtNs, ok6 := t.lookupLocked(string(triple.Target.Namespace))
	// both may be unknown independently: a wildcard grant also covers instances, which the db has never seen
	tgtInst, hasTgtInst := t.lookupLocked(string(triple.Target.Instance))
	wildcard, hasWildcard := t.lookupLocked(string(AllInstances))
	t.mutex.RUnlock()

	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
		// a string which has never been interned cannot be part of any membership or grant
		return false, true
	}

	if !hasTgtInst && !hasWildcard {
		return false, true
	}

	// the backward index stores target, relation and source, thus all members of the source are a single range
	var pivot fixedBinaryTriple
	binary.LittleEndian.PutUint32(pivot[0:4], srcNs)
	binary.LittleEndian.PutUint32(pivot[4:8], srcInst)
	binary.LittleEndian.PutUint32(pivot[8:12], member)
	binary.LittleEndian.PutUint32(pivot[12:16], memberNs)

	grant := binaryTriple{
		A:        bEntity{Namespace: bNamespace(memberNs)},
		Relation: bRelation(rel),
		B:        bEntity{Namespace: bNamespace(tgtNs)},
	}

	db.backward.Ascend(pivot, func(item fixedBinaryTriple) bool {
		if [16]byte(item[:16]) != [16]byte(pivot[:16]) {
			return false
		}

		grant.A.Instance = bInstance(binary.LittleEndian.Uint32(item[16:20]))

		if hasTgtInst {
			grant.B.Instance = bInstance(tgtInst)
			if _, found := db.forward.Get(grant.Binary()); found {
				ok = true
				return false
			}
		}

		if hasWildcard {
			grant.B.Instance = bInstance(wildcard)
			if _, found := db.forward.Get(grant.Binary()); found {
				ok = true
				return false
			}
		}

		return true
	})

	return ok, true
}

// containsViaMemberQuery is the generic implementation of [NewSourceMemberResolver], see also
// [DB.containsViaMember].
func containsViaMemberQuery(db *DB, memberNamespace Namespace, triple Triple) (bool, error) {
	q := Select().
		Where().Source().IsNamespace(memberNamespace).
		Where().Relation().Has(Member).
		Where().Target().Is(triple.Source.Namespace, triple.Source.Instance)

	for roleMemberTriple, err := range db.Query(q) {

		if err != nil {
			return false, err
		}

		// check if a role allows the requested triple
		ok, err := db.Contains(Triple{
			Source:   roleMemberTriple.Source,
			Relation: triple.Relation,
			Target:   triple.Target,
		})

		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}

		// check if a role allows the requested triple via wildcard
		ok, err = db.Contains(Triple{
			Source: Entity{
				Namespace: roleMemberTriple.Source.Namespace,
				Instance:  roleMemberTriple.Source.Instance,
			},
			Relation: triple.Relation,
			Target: Entity{
				Namespace: triple.Target.Namespace,
				Instance:  AllInstances,
			},
		})

		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}

	}

	return false, nil
}
