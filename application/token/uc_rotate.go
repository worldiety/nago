// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewRotate(mutex *sync.Mutex, repo Repository, algo user.HashAlgorithm, reverseHashLookup *concurrent.RWMap[Hash, ID]) Rotate {
	return func(subject auth.Subject, id ID) (Plaintext, error) {
		mutex.Lock()
		defer mutex.Unlock()

		if !subject.Valid() {
			return "", user.InvalidSubjectErr
		}

		optToken, err := repo.FindByID(id)
		if err != nil {
			return "", err
		}

		if optToken.IsNone() {
			return "", os.ErrNotExist
		}

		token := optToken.Unwrap()
		rotationAllowed := subject.HasResourcePermission(rebac.Namespace(repo.Name()), rebac.Instance(token.ID), PermRotate) || token.Impersonation.UnwrapOr("") == subject.ID()
		if !rotationAllowed {
			return "", user.PermissionDeniedErr
		}

		oldHash := HashString(token.TokenHash)

		plaintext := Plaintext(rand.Text())
		hBytes, err := plaintext.TokenHash(algo)
		if err != nil {
			return "", err
		}

		// security note: see uc_create.go
		hash := HashString(hBytes)

		if _, ok := reverseHashLookup.Get(hash); ok {
			return "", fmt.Errorf("generated hash collision from random token")
		}

		token.TokenHash = hBytes
		token.Algorithm = algo

		if err := repo.Save(token); err != nil {
			return "", err
		}

		// security note: delete is important to immediately disable any future token requests using the old token
		reverseHashLookup.Delete(oldHash)
		reverseHashLookup.Put(hash, token.ID)

		return plaintext, nil
	}
}
