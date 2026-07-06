// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
)

// granteeEntity converts a [Grantee] into the corresponding ReBAC source entity. The second return value is
// false if the grantee is not valid (neither or both of user/group set).
func granteeEntity(g Grantee) (rebac.Entity, bool) {
	if !g.Valid() {
		return rebac.Entity{}, false
	}

	if g.User != "" {
		return rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(g.User)}, true
	}

	return rebac.Entity{Namespace: group.Namespace, Instance: rebac.Instance(g.Group)}, true
}

// entityGrantee is the inverse of [granteeEntity]. It reports false for any source namespace which is not a
// user or group namespace.
func entityGrantee(e rebac.Entity) (Grantee, bool) {
	switch e.Namespace {
	case user.Namespace:
		return Grantee{User: user.ID(e.Instance)}, true
	case group.Namespace:
		return Grantee{Group: group.ID(e.Instance)}, true
	default:
		return Grantee{}, false
	}
}

// fileTarget builds the ReBAC target entity for the given file id.
func fileTarget(fid FID) rebac.Entity {
	return rebac.Entity{Namespace: FileNamespace, Instance: rebac.Instance(fid)}
}

// grantFilePermissions writes the ReBAC triples that grant the given permissions from src onto the file. The
// according static rules must be registered during wiring (see application/drive/cfg), otherwise rebac.DB.Put
// reports "no such rule".
func grantFilePermissions(rdb *rebac.DB, src rebac.Entity, fid FID, perms ...permission.ID) error {
	target := fileTarget(fid)
	for _, pid := range perms {
		if err := rdb.Put(rebac.Triple{Source: src, Relation: rebac.Relation(pid), Target: target}); err != nil {
			return err
		}
	}

	return nil
}

// revokeFilePermissions removes the given permission triples from src on the file. Removing a triple which does
// not exist is not an error.
func revokeFilePermissions(rdb *rebac.DB, src rebac.Entity, fid FID, perms ...permission.ID) error {
	target := fileTarget(fid)
	for _, pid := range perms {
		if err := rdb.Delete(rebac.Triple{Source: src, Relation: rebac.Relation(pid), Target: target}); err != nil {
			return err
		}
	}

	return nil
}

// revokeAllForFile removes all ReBAC triples that target the given file instance. It is best-effort cleanup on
// delete so the ReBAC store does not accumulate dangling grants for removed files.
func revokeAllForFile(rdb *rebac.DB, fid FID) error {
	return rdb.DeleteByQuery(rebac.Select().Where().Target().Is(FileNamespace, rebac.Instance(fid)))
}

// copyGrantsFromParent copies all drive ACL grants (user/group -> parent for any of the [ACLPermissions]) onto
// the child file, so that per-file grants are inherited on create - analogous to the owner/group/mode
// inheritance. This is a snapshot copy: later changes on the parent do not propagate to already created
// children.
func copyGrantsFromParent(rdb *rebac.DB, parent, child FID) error {
	if parent == "" {
		return nil
	}

	acl := make(map[permission.ID]struct{}, len(ACLPermissions))
	for _, pid := range ACLPermissions {
		acl[pid] = struct{}{}
	}

	childTarget := fileTarget(child)
	q := rebac.Select().Where().Target().Is(FileNamespace, rebac.Instance(parent))

	// collect first, then write: we must not call rdb.Put while iterating rdb.Query as the query holds the
	// db lock and Put would deadlock trying to acquire the write lock.
	var toCopy []rebac.Triple
	for triple, err := range rdb.Query(q) {
		if err != nil {
			return err
		}

		// only copy user/group sources
		if _, ok := entityGrantee(triple.Source); !ok {
			continue
		}

		// only copy known drive ACL relations
		if _, ok := acl[permission.ID(triple.Relation)]; !ok {
			continue
		}

		toCopy = append(toCopy, rebac.Triple{Source: triple.Source, Relation: triple.Relation, Target: childTarget})
	}

	for _, triple := range toCopy {
		if err := rdb.Put(triple); err != nil {
			return err
		}
	}

	return nil
}

// readFileGrants collects the direct drive ACL grants stored on the given file, grouped by grantee.
func readFileGrants(rdb *rebac.DB, fid FID) ([]FileGrant, error) {
	acl := make(map[permission.ID]struct{}, len(ACLPermissions))
	for _, pid := range ACLPermissions {
		acl[pid] = struct{}{}
	}

	// preserve a stable order of grantees as they are first encountered
	var order []rebac.Entity
	byEntity := make(map[rebac.Entity][]permission.ID)

	q := rebac.Select().Where().Target().Is(FileNamespace, rebac.Instance(fid))
	for triple, err := range rdb.Query(q) {
		if err != nil {
			return nil, err
		}

		if _, ok := entityGrantee(triple.Source); !ok {
			continue
		}

		pid := permission.ID(triple.Relation)
		if _, ok := acl[pid]; !ok {
			continue
		}

		if _, seen := byEntity[triple.Source]; !seen {
			order = append(order, triple.Source)
		}
		byEntity[triple.Source] = append(byEntity[triple.Source], pid)
	}

	grants := make([]FileGrant, 0, len(order))
	for _, e := range order {
		grantee, _ := entityGrantee(e)
		grants = append(grants, FileGrant{Grantee: grantee, Permissions: byEntity[e]})
	}

	return grants, nil
}
