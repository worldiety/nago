// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"sync"
)

func NewDelete(mutex *sync.Mutex, repo Repository) Delete {
	return func(subject auth.Subject, id ID) error {
		mutex.Lock()
		defer mutex.Unlock()

		if !subject.Valid() {
			return user.InvalidSubjectErr
		}

		optToken, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		if optToken.IsNone() {
			// security note: this allows exposing information if a token exists or not. Not sure if this
			// is a problem at all.
			return nil
		}

		token := optToken.Unwrap()
		uid := token.Impersonation.UnwrapOr("")

		allowedToDelete := subject.HasResourcePermission(repo.Name(), string(token.ID), PermDelete) || uid == subject.ID()
		if !allowedToDelete {
			return user.PermissionDeniedErr
		}

		// security note: clean up ofin-memory cached hashes and plaintext tokens are removed using the observer of
		// the repository.
		return repo.DeleteByID(id)

	}
}
