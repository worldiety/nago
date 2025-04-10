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
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewAuthenticateSubject(repo Repository, algo user.HashAlgorithm, reverseHashLookup *concurrent.RWMap[Hash, ID], subjectFromUser user.SubjectFromUser, subjectLookup *concurrent.RWMap[Plaintext, user.Subject], anonUser user.GetAnonUser) AuthenticateSubject {
	return func(plaintext Plaintext) (auth.Subject, error) {
		subj, ok := subjectLookup.Get(plaintext)
		if ok {
			// security note: we trade security (keeping all authenticated plaintext token in-memory) against
			// speed. REST APIs must be as fast as possible and this is a reasonable compromise.
			// If we would not do this, we would limit our amount of requests to a few hundred per second at best
			// because the password hash algorithm is intentionally very expensive. Note, that the subjects
			// will enable or disable themselves automatically even though we leak them infinitely.
			return subj, nil
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
			s := newSubject(token)
			subjectLookup.Put(plaintext, s)
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
		subjectLookup.Put(plaintext, usr)

		return usr, nil
	}
}
