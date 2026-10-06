// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import "go.wdy.de/nago/application/user"

type Authenticated struct {
	Session ID
	User    user.ID
}

// LoggedOut is published, after a session lost its user by a logout, also by a failed single sign-on refresh. The
// open windows of the session show it as logged out at once.
type LoggedOut struct {
	Session ID
}
