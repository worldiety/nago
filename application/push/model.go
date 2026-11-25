// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package push sends notifications to the browsers of users by Web Push. A user subscribes a browser only by an
// action, which the application offers, see package uipush. The browser then asks for the permission. Without such
// an action, nago neither registers a service worker nor asks the browser for anything. The module is enabled by
// package cfgpush.
package push

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/webpush"
	"go.wdy.de/nago/presentation/core"
)

// SubscriptionID is derived from the endpoint, so that a browser has at most one subscription, even if another
// user subscribes it later, e.g. on a shared device.
type SubscriptionID string

// NewSubscriptionID returns the identifier of the subscription with the given endpoint.
func NewSubscriptionID(endpoint string) SubscriptionID {
	sum := sha256.Sum256([]byte(endpoint))
	return SubscriptionID(base64.RawURLEncoding.EncodeToString(sum[:]))
}

// SessionHash is the hash of a session, so that a subscription of a session can be found at its logout, without
// storing the session id, which works like a password.
type SessionHash string

// NewSessionHash returns the hash of the given session.
func NewSessionHash(id session.ID) SessionHash {
	if id == "" {
		return ""
	}

	sum := sha256.Sum256([]byte("nago.push.session:" + id))
	return SessionHash(base64.RawURLEncoding.EncodeToString(sum[:]))
}

// Subscription is the subscription of a browser of a user.
type Subscription struct {
	ID       SubscriptionID `json:"id"`
	User     user.ID        `json:"user"`
	Session  SessionHash    `json:"session,omitempty"`
	Endpoint string         `json:"endpoint"`
	P256dh   string         `json:"p256dh"`
	Auth     string         `json:"auth"`
	// Origin is the origin of the application, as the browser has seen it.
	Origin    string `json:"origin,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`

	CreatedAt     time.Time `json:"createdAt"`
	LastSuccessAt time.Time `json:"lastSuccessAt,omitzero"`
}

func (s Subscription) Identity() SubscriptionID {
	return s.ID
}

// BrowserSubscription is the subscription, as the browser has returned it.
type BrowserSubscription struct {
	Endpoint  string
	P256dh    string
	Auth      string
	Origin    string
	UserAgent string
}

// Message is a notification. A browser shows at most about 4 KB of it.
type Message struct {
	// Title is required.
	Title string
	Body  string
	// Icon is optional. By default, the icon of the manifest is shown, see
	// [application.Configurator.SetManifestIcon], or the icon of the application.
	Icon core.URI
	// Tag optionally lets a newer notification with the same tag replace an older one on the device.
	Tag string
	// Route is opened by a click on the notification. An open window of the application navigates to it, otherwise
	// a new one is opened. By default, it is the start page.
	Route core.NavigationPath
	// Values are the query parameters of the route.
	Values core.Values
	// TTL is how long the push service keeps the message for a device, which is offline. It is 24 hours by default.
	// A negative value delivers the message only to devices, which are online right now.
	TTL time.Duration
	// Urgency is normal by default.
	Urgency webpush.Urgency
}

// URL returns the path and query, which a click on the notification opens.
func (m Message) URL() string {
	u := url.URL{Path: "/" + strings.TrimPrefix(string(m.Route), "/")}
	if len(m.Values) > 0 {
		q := url.Values{}
		for k, v := range m.Values {
			q.Set(k, v)
		}

		u.RawQuery = q.Encode()
	}

	return u.String()
}

// Recipients are the users, which get a message. A user, who is included more than once, gets it only once.
type Recipients struct {
	Users  []user.ID
	Groups []group.ID
	Roles  []role.ID
}

// ToUser returns the recipients of a single user.
func ToUser(id ...user.ID) Recipients {
	return Recipients{Users: id}
}

// ToGroup returns all members of the groups as recipients.
func ToGroup(id ...group.ID) Recipients {
	return Recipients{Groups: id}
}

// ToRole returns all members of the roles as recipients.
func ToRole(id ...role.ID) Recipients {
	return Recipients{Roles: id}
}

// Report tells what happened to the subscriptions of the recipients.
type Report struct {
	// Users is the number of distinct recipients.
	Users int
	// Sent is the number of subscriptions, whose push service has accepted the message.
	Sent int
	// Removed is the number of subscriptions, which the push services did not know anymore and which have been
	// deleted.
	Removed int
	// Failed is the number of subscriptions, whose push service has rejected the message or was unreachable.
	Failed int
}

// ServiceWorkerPath is the path of the service worker of nago, which receives the push messages and opens the
// application by a click on a notification. It exists only, if the module is enabled.
const ServiceWorkerPath = "/api/nago/v1/push/sw.js"
