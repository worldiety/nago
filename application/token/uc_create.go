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
	"sync"
	"time"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewCreate(mutex *sync.Mutex, repo Repository, algo user.HashAlgorithm, reverseHashLookup *concurrent.RWMap[Hash, ID]) Create {
	return func(subject auth.Subject, cdata CreationData) (ID, Plaintext, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		// security note: we do not use a salt here:
		// * our implementation ensures that tokens are always unique to allow efficient inverse lookups.
		//   Collisions between instances are astronomical.
		// * we cannot increase complexity or entropy beyond our secure random source
		// * we must be efficient to pre-calculate the inverse hash lookup table, which is not possible
		//   with salts (which is exactly the point of a salt).

		// security note: our token has at least 16 bytes of entropy, which is ever returned once and never stored
		plaintext := cdata.Plaintext
		if plaintext == "" {
			plaintext = Plaintext(rand.Text())
		}

		if len(plaintext) < 16 {
			return "", "", fmt.Errorf("plaintext is too short: %s", plaintext)
		}

		hBytes, err := plaintext.TokenHash(algo)
		if err != nil {
			return "", "", err
		}

		hash := HashString(hBytes)

		// security note: this is only valid as long as we do not change the hash algorithm.
		// If the way we treat these hashes becomes invalid, we must remove/revoke/rotate all of them at once, to
		// avoid potential attacks on broken algorithms or collisions across the output of different algorithms.
		if _, ok := reverseHashLookup.Get(hash); ok {
			return "", "", fmt.Errorf("generated hash collision from random token")
		}

		token := Token{
			ID:          data.RandIdent[ID](),
			Name:        cdata.Name,
			Description: cdata.Description,
			Algorithm:   algo,
			TokenHash:   hBytes,
			CreatedAt:   time.Now(),
			ValidUntil:  cdata.ValidUntil,
		}

		optToken, err := repo.FindByID(token.ID)
		if err != nil {
			return "", "", err
		}

		if optToken.IsSome() {
			return "", "", fmt.Errorf("token already exists")
		}

		if err := repo.Save(token); err != nil {
			return "", "", err
		}

		reverseHashLookup.Put(hash, token.ID)

		return token.ID, plaintext, nil
	}
}
