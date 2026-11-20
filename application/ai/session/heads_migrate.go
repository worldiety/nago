// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

// migrateWithoutHeads starts the build for a store of sessions, which has never had heads, because a former version of
// nago has written it.
//
// This file can be removed, once every instance has run a version with [Heads]: the call in heads.start goes with it.
// Afterwards, a store of a former version is never built, and its searches read every session. The build after an
// interrupted build or a restore stays.
func migrateWithoutHeads(h *heads) {
	h.kick()
}
