// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

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
}
