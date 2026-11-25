// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package cfgpush enables the push notifications of package push.
package cfgpush

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/push"
	uipush "go.wdy.de/nago/application/push/ui"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/webpush"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/web/vuejs"
)

type Options struct {
	// Subject is the contact of the application for the push services, a mailto: or https: URL. By default, the
	// origin of the application is used, which the browser has reported at its subscription.
	Subject string

	// KeepOnLogout keeps the subscriptions of a session at its logout. By default, a browser gets no notifications
	// anymore, after its user has logged out, because another person may use it next.
	KeepOnLogout bool

	// AccountSection adds a switch for the notifications of the browser to the account view of nago. By default,
	// the application decides, where to offer the notifications, see [uipush.Card] and [uipush.Toggle].
	AccountSection bool

	// AllowEndpoint optionally replaces the allowed push services, see [push.KnownPushService].
	AllowEndpoint func(u *url.URL) bool

	// MaxPerUser limits the number of browsers per user, default 32.
	MaxPerUser int

	// Client optionally replaces the http client, which posts to the push services.
	Client *http.Client
}

type Management struct {
	UseCases push.UseCases
}

// AccountEntry is the entry of [Options.AccountSection] in the account view.
const AccountEntry uiuser.AccountEntryID = "nago-push"

// Enable installs the push notifications. It creates the VAPID key at the first start and keeps it encrypted by the
// master key. Nothing is shown to the user and the browser asks nothing, until the application offers the
// notifications, e.g. by [uipush.Card]. It requires the user management.
func Enable(cfg *application.Configurator, opts Options) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	users, err := cfg.UserManagement()
	if err != nil {
		return Management{}, fmt.Errorf("push requires the user management: %w", err)
	}

	rdb, err := cfg.RDB()
	if err != nil {
		return Management{}, fmt.Errorf("push requires the relations: %w", err)
	}

	keys, err := encryptedRepository[vapidKey, vapidKeyID](cfg, "nago.push.vapid")
	if err != nil {
		return Management{}, fmt.Errorf("cannot get the store of the push key: %w", err)
	}

	key, err := loadOrCreateKey(keys)
	if err != nil {
		return Management{}, err
	}

	// the endpoints and auth secrets tell, which browsers belong to whom, so they are encrypted like the sessions
	repo, err := encryptedRepository[push.Subscription, push.SubscriptionID](cfg, "nago.push.subscription")
	if err != nil {
		return Management{}, fmt.Errorf("cannot get the store of the push subscriptions: %w", err)
	}

	uc := push.NewUseCases(push.Options{
		Key:           key,
		Subject:       opts.Subject,
		DefaultIcon:   func() string { return string(cfg.ManifestIcon()) },
		AllowEndpoint: opts.AllowEndpoint,
		MaxPerUser:    opts.MaxPerUser,
		Client:        opts.Client,
	}, push.Deps{
		Repository: repo,
		RDB:        rdb,
		UserExists: func(uid user.ID) bool {
			optUser, err := users.UseCases.FindByID(user.SU(), uid)
			return err == nil && optUser.IsSome() && optUser.Unwrap().Enabled()
		},
	})

	if !opts.KeepOnLogout {
		events.SubscribeFor[session.LoggedOut](cfg.EventBus(), func(evt session.LoggedOut) {
			if err := uc.DeleteBySession(evt.Session); err != nil {
				slog.Error("push: cannot delete the subscriptions of a session", "err", err)
			}
		})
	}

	events.SubscribeFor[user.Deleted](cfg.EventBus(), func(evt user.Deleted) {
		if err := uc.DeleteByUser(evt.ID); err != nil {
			slog.Error("push: cannot delete the subscriptions of a deleted user", "user", evt.ID, "err", err)
		}
	})

	cfg.HandleMethod(http.MethodGet, push.ServiceWorkerPath, serveServiceWorker)

	if opts.AccountSection {
		cfg.AddAccountSections(func(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
			if !wnd.Subject().Valid() {
				return sections
			}

			return uiuser.InsertAccountSection(sections, uiuser.SectionAccount, uiuser.AccountSectionModel{
				ID:    "nago-push",
				Title: uipush.StrTitle.Get(wnd),
				Entries: []uiuser.AccountEntryModel{
					{ID: AccountEntry, View: uipush.Toggle(wnd)},
				},
			})
		})
	}

	management = Management{UseCases: uc}
	cfg.AddContextValue(core.ContextValue("nago.push.usecases", uc))
	cfg.AddContextValue(core.ContextValue("nago.push.management", management))
	slog.Info("installed push module")

	return management, nil
}

// serveServiceWorker serves the service worker for the scope of the whole application. It must be revalidated,
// so that a new version of nago replaces it soon.
func serveServiceWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Service-Worker-Allowed", "/")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "push.js", time.Time{}, bytes.NewReader(vuejs.PushServiceWorker))
}

// encryptedRepository returns a repository, whose entities are encrypted by the master key, like those of the
// sessions and secrets.
func encryptedRepository[A data.Aggregate[ID], ID data.IDType](cfg *application.Configurator, bucketName string) (data.Repository[A, ID], error) {
	masterKey, err := cfg.MasterKey()
	if err != nil {
		return nil, fmt.Errorf("push requires the master key: %w", err)
	}

	store, err := cfg.EntityStore(bucketName)
	if err != nil {
		return nil, err
	}

	return json.NewSloppyJSONRepository[A, ID](crypto.NewBlobStore(store, masterKey)), nil
}

type vapidKeyID string

// vapidKey is the private VAPID key. There is only one, see [currentVAPIDKey].
type vapidKey struct {
	ID         vapidKeyID `json:"id"`
	PrivateKey string     `json:"privateKey"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (k vapidKey) Identity() vapidKeyID {
	return k.ID
}

const currentVAPIDKey vapidKeyID = "vapid"

// loadOrCreateKey reads the VAPID key. A new key would make all subscriptions useless, so a key, which cannot be
// read, e.g. after the master key has changed, is an error and not replaced.
func loadOrCreateKey(repo data.Repository[vapidKey, vapidKeyID]) (*ecdsa.PrivateKey, error) {
	optKey, err := repo.FindByID(currentVAPIDKey)
	if err != nil {
		return nil, fmt.Errorf("cannot read the push key: %w", err)
	}

	if optKey.IsSome() {
		key, err := webpush.ParsePrivateKey(optKey.Unwrap().PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("cannot decode the push key: %w", err)
		}

		return key, nil
	}

	key, err := webpush.GenerateKey()
	if err != nil {
		return nil, err
	}

	encoded, err := webpush.MarshalPrivateKey(key)
	if err != nil {
		return nil, err
	}

	if err := repo.Save(vapidKey{ID: currentVAPIDKey, PrivateKey: encoded, CreatedAt: time.Now()}); err != nil {
		return nil, fmt.Errorf("cannot store the push key: %w", err)
	}

	slog.Info("push: created a new VAPID key")
	return key, nil
}
