// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
)

// InstancePermissions are the per-session permissions granted to a session's creator so that a user can
// access exactly their own sessions even without any global session permission. Create is intentionally NOT
// part of this list: creating is a global capability (typically assigned to an IAM group), not something one
// is granted on a specific instance.
//
// The wiring (application/ai/cfg) registers a ReBAC static rule user -> session for each of these plus the
// [rebac.Owner] relation, which is required before [grantOwner] may write the corresponding triples.
var InstancePermissions = []permission.ID{
	PermFindByID,
	PermFindAll,
	PermAppend,
	PermRename,
	PermDelete,
}

// grantOwner writes the ReBAC triples that make the given user the owner of the session instance and grant
// them all per-instance permissions. It is called on create. The corresponding static rules must be
// registered during wiring (see application/ai/cfg), otherwise rdb.Put reports "no such rule".
func grantOwner(rdb *rebac.DB, uid user.ID, sid ID) error {
	src := rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(uid)}
	target := rebac.Entity{Namespace: Namespace, Instance: rebac.Instance(sid)}

	if err := rdb.Put(rebac.Triple{Source: src, Relation: rebac.Owner, Target: target}); err != nil {
		return err
	}

	for _, pid := range InstancePermissions {
		if err := rdb.Put(rebac.Triple{Source: src, Relation: rebac.Relation(pid), Target: target}); err != nil {
			return err
		}
	}

	return nil
}

// revokeInstance removes all ReBAC triples that target the given session instance. It is best-effort cleanup
// on delete so the ReBAC store does not accumulate dangling grants for removed sessions.
func revokeInstance(rdb *rebac.DB, sid ID) error {
	return rdb.DeleteByQuery(rebac.Select().Where().Target().Is(Namespace, rebacInstance(sid)))
}

// rebacInstance converts a session id into a ReBAC instance identifier.
func rebacInstance(sid ID) rebac.Instance {
	return rebac.Instance(sid)
}
