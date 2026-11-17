// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"context"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

// authenticated is the subject, which [AuthenticateSubject] found for a plaintext, with the token, to which the
// plaintext belongs. The subject of an impersonating token is the user, so its ID does not name the token.
type authenticated struct {
	token   ID
	subject user.Subject
}

func NewAuthenticateSubject(
	ctx context.Context,
	mutex *sync.Mutex,
	repo Repository,
	algo user.HashAlgorithm,
	reverseHashLookup *concurrent.RWMap[Hash, ID],
	subjectFromUser user.SubjectFromUser,
	subjectLookup *concurrent.RWMap[Plaintext, authenticated],
	anonUser user.GetAnonUser,
	findRoleByID role.FindByID,
	rdb *rebac.DB,
) AuthenticateSubject {
	return func(plaintext Plaintext) (auth.Subject, error) {
		cached, ok := subjectLookup.Get(plaintext)
		if ok {
			// security note: we trade security (keeping all authenticated plaintext token in-memory) against
			// speed. REST APIs must be as fast as possible and this is a reasonable compromise.
			// If we would not do this, we would limit our amount of requests to a few hundred per second at best
			// because the password hash algorithm is intentionally very expensive. Note, that the subjects
			// will enable or disable themselves automatically even though we leak them infinitely.
			return cached.subject, nil
		}

		// security note: we currently expect that all hash algorithms are of the same and given kind. Otherwise,
		// we will reject them. We don't try to perform a kind of fallback here.
		// However, we are still prone to DoS attacks causing massive loads by invoking with invalid tokens but at
		// least we are still O(1).
		hbytes, err := plaintext.TokenHash(algo)
		if err != nil {
			return nil, err
		}

		hash := HashString(hbytes)

		// security note: the expensive hash is computed above without the lock, but the lookup and the caching
		// below hold it, so that a rotation or a deletion of the token cannot happen in between and leave the
		// old plaintext in the lookup of the subjects
		mutex.Lock()
		defer mutex.Unlock()

		tid, ok := reverseHashLookup.Get(hash)
		if !ok {
			//
			return anonUser(), nil
		}

		// security note: we bypass the constant time comparison of the argon2id (or whatever algorithm)
		// but due to the O(1) lookup mechanic above we already destroyed any constant protection anyway.
		optToken, err := repo.FindByID(tid)
		if err != nil {
			return nil, err
		}

		if optToken.IsNone() {
			return anonUser(), nil
		}

		token := optToken.Unwrap()

		if token.Impersonation.IsNone() {
			s := newSubject(ctx, findRoleByID, repo, token, rdb)
			subjectLookup.Put(plaintext, authenticated{token: tid, subject: s})
			return s, nil
		}

		uid := token.Impersonation.Unwrap()
		optUsr, err := subjectFromUser(user.SU(), uid)
		if err != nil {
			return nil, err
		}

		if optUsr.IsNone() {
			return anonUser(), nil
		}

		usr := optUsr.Unwrap()
		subjectLookup.Put(plaintext, authenticated{token: tid, subject: usr})

		return usr, nil
	}
}
