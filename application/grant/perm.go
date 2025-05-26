// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package grant

import "go.wdy.de/nago/application/permission"

var (
	PermGrant       = permission.Declare[Grant]("nago.grant.grant", "Grant permissions to others", "A user with that permission assigned can grant permissions to other users.")
	PermListGranted = permission.Declare[ListGranted]("nago.grant.listgranted", "List granted users for resource", "A user with that permission assigned can list other users which have granted permissions on a specific resource.")
	PermListGrants  = permission.Declare[ListGrants]("nago.grant.listgrants", "List permissions for a users resource", "A user with that permission assigned can list granted permissions for specific user and resource.")
)
