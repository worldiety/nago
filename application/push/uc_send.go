// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/webpush"
)

// payload is read by the service worker of nago, see web/vuejs/serviceworker/push.js.
type payload struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Tag   string `json:"tag,omitempty"`
	URL   string `json:"url"`
}

// deliver sends the message to the subscriptions of the given users.
type deliver func(users map[user.ID]struct{}, msg Message) (Report, error)

func NewSend(deps Deps, deliver deliver) Send {
	return func(subject auth.Subject, to Recipients, msg Message) (Report, error) {
		if err := subject.Audit(PermSend); err != nil {
			return Report{}, err
		}

		users, err := resolve(deps, to)
		if err != nil {
			return Report{}, err
		}

		return deliver(users, msg)
	}
}

func NewSendToMe(deliver deliver) SendToMe {
	return func(subject auth.Subject, msg Message) (Report, error) {
		if !subject.Valid() {
			return Report{}, user.InvalidSubjectErr
		}

		return deliver(map[user.ID]struct{}{subject.ID(): {}}, msg)
	}
}

// resolve returns the distinct users of the recipients.
func resolve(deps Deps, to Recipients) (map[user.ID]struct{}, error) {
	users := map[user.ID]struct{}{}
	for _, uid := range to.Users {
		users[uid] = struct{}{}
	}

	members := func(ns rebac.Namespace, instance rebac.Instance) error {
		if deps.RDB == nil {
			return fmt.Errorf("cannot resolve the members of %s without relations", ns)
		}

		it := deps.RDB.Query(rebac.Select().
			Where().Source().Is(ns, instance).
			Where().Relation().Has(rebac.Member).
			Where().Target().IsNamespace(user.Namespace),
		)

		for triple, err := range it {
			if err != nil {
				return err
			}

			users[user.ID(triple.Target.Instance)] = struct{}{}
		}

		return nil
	}

	for _, gid := range to.Groups {
		if err := members(group.Namespace, rebac.Instance(gid)); err != nil {
			return nil, err
		}
	}

	for _, rid := range to.Roles {
		if err := members(role.Namespace, rebac.Instance(rid)); err != nil {
			return nil, err
		}
	}

	for uid := range users {
		if uid == "" || !deps.UserExists(uid) {
			delete(users, uid)
		}
	}

	return users, nil
}

func newDeliver(opts Options, repo Repository) deliver {
	return func(users map[user.ID]struct{}, msg Message) (Report, error) {
		report := Report{Users: len(users)}
		if opts.Key == nil {
			return report, fmt.Errorf("push has no VAPID key")
		}

		if msg.Title == "" {
			return report, fmt.Errorf("a push message requires a title")
		}

		icon := string(msg.Icon)
		if icon == "" {
			icon = opts.DefaultIcon()
		}

		buf, err := json.Marshal(payload{
			Title: msg.Title,
			Body:  msg.Body,
			Icon:  icon,
			Tag:   msg.Tag,
			URL:   msg.URL(),
		})
		if err != nil {
			return report, err
		}

		if len(buf) > webpush.MaxPayloadSize {
			return report, fmt.Errorf("push message has %d bytes but at most %d bytes are possible", len(buf), webpush.MaxPayloadSize)
		}

		ttl := msg.TTL
		switch {
		case ttl == 0:
			ttl = 24 * time.Hour
		case ttl < 0:
			ttl = 0
		}

		var subs []Subscription
		for sub, err := range repo.All() {
			if err != nil {
				return report, err
			}

			if _, ok := users[sub.User]; ok {
				subs = append(subs, sub)
			}
		}

		var mutex sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, opts.Parallel)
		for _, sub := range subs {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				err := webpush.Send(context.Background(), webpush.Subscription{
					Endpoint: sub.Endpoint,
					Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
				}, buf, webpush.Options{
					Key:     opts.Key,
					Subject: vapidSubject(opts.Subject, sub.Origin),
					TTL:     ttl,
					Urgency: msg.Urgency,
					Client:  opts.Client,
					Now:     opts.Now,
				})

				mutex.Lock()
				defer mutex.Unlock()

				switch {
				case err == nil:
					report.Sent++
					// a logout may have deleted the subscription meanwhile, which must not come back
					if now := opts.Now(); now.Sub(sub.LastSuccessAt) > time.Hour {
						if optCur, err := repo.FindByID(sub.ID); err == nil && optCur.IsSome() && optCur.Unwrap().User == sub.User {
							cur := optCur.Unwrap()
							cur.LastSuccessAt = now
							if err := repo.Save(cur); err != nil {
								slog.Error("cannot update push subscription", "id", sub.ID, "err", err)
							}
						}
					}
				case errors.Is(err, webpush.ErrGone):
					report.Removed++
					if err := repo.DeleteByID(sub.ID); err != nil {
						slog.Error("cannot delete push subscription", "id", sub.ID, "err", err)
					}
				default:
					report.Failed++
					slog.Warn("push service did not accept the message", "id", sub.ID, "user", sub.User, "err", err)
				}
			}()
		}

		wg.Wait()

		return report, nil
	}
}

// vapidSubject returns the configured contact or the origin of the application, as the browser has seen it.
func vapidSubject(configured string, origin string) string {
	if configured != "" {
		return configured
	}

	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return ""
	}

	return u.Scheme + "://" + u.Host
}
