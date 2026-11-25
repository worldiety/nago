// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package push

import (
	"crypto/ecdsa"
	"fmt"
	"iter"
	"net/url"
	"slices"
	"strings"
	"sync"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/webpush"
)

// registerMutex serializes the registrations, so that the limit of subscriptions per user holds.
var registerMutex sync.Mutex

func NewRegister(opts Options, repo Repository) Register {
	return func(subject auth.Subject, sid session.ID, sub BrowserSubscription) (Subscription, error) {
		if !subject.Valid() {
			return Subscription{}, user.InvalidSubjectErr
		}

		u, err := url.Parse(sub.Endpoint)
		if err != nil || u.Host == "" {
			return Subscription{}, fmt.Errorf("invalid endpoint of the push service")
		}

		if !opts.AllowEndpoint(u) {
			return Subscription{}, fmt.Errorf("the push service %s is not allowed", u.Host)
		}

		// encrypt a message once, so that invalid keys are rejected now and not at each message
		if _, err := webpush.Encrypt(webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth}, nil); err != nil {
			return Subscription{}, fmt.Errorf("invalid keys of the subscription: %w", err)
		}

		registerMutex.Lock()
		defer registerMutex.Unlock()

		id := NewSubscriptionID(sub.Endpoint)
		s := Subscription{
			ID:        id,
			User:      subject.ID(),
			Session:   NewSessionHash(sid),
			Endpoint:  sub.Endpoint,
			P256dh:    sub.P256dh,
			Auth:      sub.Auth,
			Origin:    limit(sub.Origin, 256),
			UserAgent: limit(sub.UserAgent, 512),
			CreatedAt: opts.Now(),
		}

		if optOld, err := repo.FindByID(id); err != nil {
			return Subscription{}, err
		} else if optOld.IsSome() && optOld.Unwrap().User == s.User {
			s.CreatedAt = optOld.Unwrap().CreatedAt
			s.LastSuccessAt = optOld.Unwrap().LastSuccessAt
		}

		if err := repo.Save(s); err != nil {
			return Subscription{}, err
		}

		// keep only the newest subscriptions of the user
		var mine []Subscription
		for other, err := range repo.All() {
			if err != nil {
				return Subscription{}, err
			}

			if other.User == s.User {
				mine = append(mine, other)
			}
		}

		if len(mine) > opts.MaxPerUser {
			slices.SortFunc(mine, func(a, b Subscription) int {
				return b.CreatedAt.Compare(a.CreatedAt)
			})

			for _, old := range mine[opts.MaxPerUser:] {
				if old.ID == s.ID {
					continue
				}

				if err := repo.DeleteByID(old.ID); err != nil {
					return Subscription{}, err
				}
			}
		}

		return s, nil
	}
}

func limit(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return strings.ToValidUTF8(s[:n], "")
}

func NewUnregister(repo Repository) Unregister {
	return func(subject auth.Subject, endpoint string) error {
		if !subject.Valid() {
			return user.InvalidSubjectErr
		}

		return deleteIfMine(repo, subject, NewSubscriptionID(endpoint))
	}
}

func deleteIfMine(repo Repository, subject auth.Subject, id SubscriptionID) error {
	optSub, err := repo.FindByID(id)
	if err != nil {
		return err
	}

	if optSub.IsNone() || optSub.Unwrap().User != subject.ID() {
		return nil
	}

	return repo.DeleteByID(id)
}

func NewIsRegistered(repo Repository) IsRegistered {
	return func(subject auth.Subject, endpoint string) (bool, error) {
		if !subject.Valid() || endpoint == "" {
			return false, nil
		}

		optSub, err := repo.FindByID(NewSubscriptionID(endpoint))
		if err != nil {
			return false, err
		}

		return optSub.IsSome() && optSub.Unwrap().User == subject.ID(), nil
	}
}

func NewFindMySubscriptions(repo Repository) FindMySubscriptions {
	return func(subject auth.Subject) iter.Seq2[Subscription, error] {
		return func(yield func(Subscription, error) bool) {
			if !subject.Valid() {
				yield(Subscription{}, user.InvalidSubjectErr)
				return
			}

			for sub, err := range repo.All() {
				if err != nil {
					yield(Subscription{}, err)
					return
				}

				if sub.User != subject.ID() {
					continue
				}

				if !yield(sub, nil) {
					return
				}
			}
		}
	}
}

func NewDeleteMySubscription(repo Repository) DeleteMySubscription {
	return func(subject auth.Subject, id SubscriptionID) error {
		if !subject.Valid() {
			return user.InvalidSubjectErr
		}

		return deleteIfMine(repo, subject, id)
	}
}

func newDeleteBySession(repo Repository) func(sid session.ID) error {
	return func(sid session.ID) error {
		hash := NewSessionHash(sid)
		if hash == "" {
			return nil
		}

		return repo.Delete(func(s Subscription) (bool, error) {
			return s.Session == hash, nil
		})
	}
}

func newDeleteByUser(repo Repository) func(uid user.ID) error {
	return func(uid user.ID) error {
		return repo.Delete(func(s Subscription) (bool, error) {
			return s.User == uid, nil
		})
	}
}

func publicKeyOf(key *ecdsa.PrivateKey) (string, error) {
	return webpush.PublicKey(key)
}
