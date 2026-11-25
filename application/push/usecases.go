// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package push

import (
	"crypto/ecdsa"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

var (
	PermSend = permission.Declare[Send]("nago.push.send", "Push-Benachrichtigungen senden", "Träger dieser Berechtigung können Push-Benachrichtigungen an beliebige Nutzer, Gruppen und Rollen senden.")
)

// Repository stores the subscriptions.
type Repository = data.Repository[Subscription, SubscriptionID]

// Register stores the subscription of the browser for the user of the subject and the given session. A subscription
// of the same browser is replaced, even if it belonged to another user before, e.g. on a shared device.
type Register func(subject auth.Subject, sid session.ID, sub BrowserSubscription) (Subscription, error)

// Unregister deletes the subscription with the endpoint, if it belongs to the subject. An unknown endpoint is no
// error.
type Unregister func(subject auth.Subject, endpoint string) error

// IsRegistered tells, whether the subscription with the endpoint belongs to the subject.
type IsRegistered func(subject auth.Subject, endpoint string) (bool, error)

// FindMySubscriptions returns the subscriptions of the subject, i.e. the browsers, which get its notifications.
type FindMySubscriptions func(subject auth.Subject) iter.Seq2[Subscription, error]

// DeleteMySubscription deletes a subscription of the subject.
type DeleteMySubscription func(subject auth.Subject, id SubscriptionID) error

// Send sends the message to all browsers of the recipients. It requires [PermSend] and blocks, until all push
// services have answered. A subscription, which a push service does not know anymore, is deleted.
type Send func(subject auth.Subject, to Recipients, msg Message) (Report, error)

// SendToMe sends the message to all browsers of the subject, e.g. to try the notifications. It requires no
// permission but a user.
type SendToMe func(subject auth.Subject, msg Message) (Report, error)

// PublicKey returns the public VAPID key, which a browser requires to subscribe, as base64url.
type PublicKey func() string

type UseCases struct {
	Register             Register
	Unregister           Unregister
	IsRegistered         IsRegistered
	FindMySubscriptions  FindMySubscriptions
	DeleteMySubscription DeleteMySubscription
	Send                 Send
	SendToMe             SendToMe
	PublicKey            PublicKey

	// DeleteBySession deletes the subscriptions of a session, e.g. after its logout.
	DeleteBySession func(sid session.ID) error
	// DeleteByUser deletes the subscriptions of a user, e.g. after the user has been deleted.
	DeleteByUser func(uid user.ID) error
}

// Options of [NewUseCases].
type Options struct {
	// Key is the private VAPID key.
	Key *ecdsa.PrivateKey
	// Subject is the contact of the sender, a mailto: or https: URL, which the push services may use, if something
	// goes wrong. By default, the https origin of the application is used, which the browser has reported at the
	// subscription. A local origin, like during the development on localhost, is left out, because the push service
	// of Apple rejects it.
	Subject string
	// DefaultIcon is shown by a notification without icon.
	DefaultIcon func() string
	// AllowEndpoint decides, which push services may be used. The endpoints come from the browser and thus from
	// the user, so that an arbitrary endpoint would let the server post to any address. By default, only https
	// endpoints of the push services of Google, Mozilla, Apple and Microsoft are allowed, see [KnownPushService].
	AllowEndpoint func(u *url.URL) bool
	// MaxPerUser limits the number of subscriptions per user. The oldest one is deleted, if a user subscribes
	// another browser. Default is 32.
	MaxPerUser int
	// Client sends the messages. By default, it has a timeout of 30 seconds.
	Client *http.Client
	// Parallel is the number of push services, which are posted to at the same time. Default is 8.
	Parallel int
	// Now is optional and [time.Now] by default.
	Now func() time.Time
}

// Deps of [NewUseCases].
type Deps struct {
	Repository Repository
	RDB        *rebac.DB
	// UserExists optionally filters recipients, which do not exist anymore or which are disabled.
	UserExists func(uid user.ID) bool
}

// pushServiceHosts are the hosts of the push services of the common browsers. A subdomain is allowed as well.
var pushServiceHosts = []string{
	"fcm.googleapis.com",          // Chrome, Edge on Android, Opera, Samsung Internet and other Chromium browsers
	"android.googleapis.com",      // former endpoints of Chrome
	"push.services.mozilla.com",   // Firefox, e.g. updates.push.services.mozilla.com
	"push.apple.com",              // Safari, e.g. web.push.apple.com
	"notify.windows.com",          // Edge on Windows, e.g. wns2-par02p.notify.windows.com
	"push.services.microsoft.com", // reserved for Edge
}

// KnownPushService reports, whether the URL is an https endpoint of a push service of Google, Mozilla, Apple or
// Microsoft.
func KnownPushService(u *url.URL) bool {
	if u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	for _, known := range pushServiceHosts {
		if host == known || strings.HasSuffix(host, "."+known) {
			return true
		}
	}

	return false
}

func NewUseCases(opts Options, deps Deps) UseCases {
	if opts.AllowEndpoint == nil {
		opts.AllowEndpoint = KnownPushService
	}

	if opts.MaxPerUser <= 0 {
		opts.MaxPerUser = 32
	}

	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 30 * time.Second}
	}

	if opts.Parallel <= 0 {
		opts.Parallel = 8
	}

	if opts.Now == nil {
		opts.Now = time.Now
	}

	if opts.DefaultIcon == nil {
		opts.DefaultIcon = func() string { return "" }
	}

	if deps.UserExists == nil {
		deps.UserExists = func(uid user.ID) bool { return true }
	}

	var publicKey string
	if opts.Key != nil {
		publicKey, _ = publicKeyOf(opts.Key)
	}

	deliver := newDeliver(opts, deps.Repository)

	return UseCases{
		Register:             NewRegister(opts, deps.Repository),
		Unregister:           NewUnregister(deps.Repository),
		IsRegistered:         NewIsRegistered(deps.Repository),
		FindMySubscriptions:  NewFindMySubscriptions(deps.Repository),
		DeleteMySubscription: NewDeleteMySubscription(deps.Repository),
		Send:                 NewSend(deps, deliver),
		SendToMe:             NewSendToMe(deliver),
		PublicKey:            func() string { return publicKey },
		DeleteBySession:      newDeleteBySession(deps.Repository),
		DeleteByUser:         newDeleteByUser(deps.Repository),
	}
}
