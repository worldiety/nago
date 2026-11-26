// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// embeddedPrefix starts the id of a session, which nago created for a page in a frame of another site, see
// [NewEmbeddedID].
const embeddedPrefix = "embedded-"

// ErrEmbedded is the error of a login into a session of an embedded page, see [NewEmbeddedID].
var ErrEmbedded = fmt.Errorf("a session of a page in a frame of another site cannot log in: %w", os.ErrPermission)

// NewEmbeddedID returns a new random id of a session for a page, which another site shows in a frame. Such a session
// never has a user: [Login], [LoginUser] and [StartNLSFlow] refuse it with [ErrEmbedded], and a user, which has been
// stored for it anyway, is not reported. Otherwise, another site could show a page of a logged-in user in its frame and trick the
// user into clicks (clickjacking). The id tells it, thus it needs no record, which a restart could lose.
func NewEmbeddedID() ID {
	var tmp [32]byte
	_, _ = rand.Read(tmp[:]) // never fails, see crypto/rand.Read
	return ID(embeddedPrefix + hex.EncodeToString(tmp[:]))
}

// Embedded reports whether the id belongs to a session of an embedded page, see [NewEmbeddedID].
func (id ID) Embedded() bool {
	return strings.HasPrefix(string(id), embeddedPrefix)
}
