// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package gostrictnet couples the connections of the programs of package lib/net with the secrets of type
// [secret.HTTP]. Importing it offers the type on the pages of the secrets.
package gostrictnet

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
)

var _ = enum.Variant[secret.Credentials, secret.HTTP](
	enum.Rename[secret.HTTP]("nago.secret.http"),
)

// The kinds of the login of a connection, see [secret.HTTP].
const (
	AuthNone   = "none"
	AuthBearer = "bearer"
	AuthBasic  = "basic"
	AuthHeader = "header"
)

// Resolver returns the resolver of the connections of the secrets of type [secret.HTTP], which are shared with the
// system group, like the SMTP servers of the mails. So a user of a program needs no access to a token and every program
// of the application may use every such connection: an application, which separates its programs, needs a resolver of
// its own. It searches a connection by the exact name, because a program must never reach another system.
func Resolver(sys user.SysUser, secrets secret.FindGroupSecrets) nethost.Resolver {
	return func(name string) (nethost.Connection, error) {
		var found *secret.HTTP
		for c, err := range secret.FindGroupCredentialsForType[secret.HTTP](sys(), secrets, group.System) {
			if err != nil {
				return nethost.Connection{}, err
			}
			if c.Name != name {
				continue
			}
			if found != nil {
				return nethost.Connection{}, fmt.Errorf("the name of the connection %q is ambiguous", name)
			}
			found = &c
		}
		if found == nil {
			return nethost.Connection{}, fmt.Errorf("no connection %q is shared with the system group", name)
		}
		return Connection(*found)
	}
}

// Connection returns the connection of the secret.
func Connection(c secret.HTTP) (nethost.Connection, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nethost.Connection{}, errors.New("the base URL of the connection is invalid")
	}
	if base.Host == "" || base.User != nil || (base.Scheme != "http" && base.Scheme != "https") {
		return nethost.Connection{}, errors.New("the base URL of the connection must be http or https, have a host and no user")
	}

	switch c.Auth {
	case "", AuthNone:
		return nethost.Connection{BaseURL: base}, nil
	case AuthBearer:
		value := "Bearer " + c.Token
		return nethost.Connection{BaseURL: base, Decorate: setHeader("Authorization", value), Secrets: []string{c.Token}}, nil
	case AuthBasic:
		encoded := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Token))
		return nethost.Connection{BaseURL: base, Decorate: setHeader("Authorization", "Basic "+encoded), Secrets: []string{c.Token, encoded}}, nil
	case AuthHeader:
		if c.HeaderName == "" {
			return nethost.Connection{}, errors.New("the connection needs the name of its header")
		}
		return nethost.Connection{BaseURL: base, Decorate: setHeader(c.HeaderName, c.Token), Secrets: []string{c.Token}}, nil
	}
	return nethost.Connection{}, fmt.Errorf("the connection has the unknown login %q", c.Auth)
}

// setHeader returns a decorator, which sets the header of each request.
func setHeader(name, value string) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(r *http.Request) (*http.Response, error) {
			r = r.Clone(r.Context())
			r.Header.Set(name, value)
			return next.RoundTrip(r)
		})
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
