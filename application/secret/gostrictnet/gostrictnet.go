// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package gostrictnet couples the connections of the programs of package lib/net with the secrets of type
// [secret.KeyValues]. Importing it offers the type on the pages of the secrets.
package gostrictnet

import (
	"cmp"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
)

var _ = enum.Variant[secret.Credentials, secret.KeyValues](
	enum.Rename[secret.KeyValues]("nago.secret.keyvalues"),
)

// Resolver returns the resolver of the connections of the secrets of type [secret.KeyValues], which are shared with the
// system group, like the SMTP servers of the mails. So a user of a program needs no access to a token and every program
// of the application may use every such connection: an application, which separates its programs, needs a resolver of
// its own. It searches a connection by the exact name, because a program must never reach another system, and a name
// which is used twice is ambiguous.
func Resolver(sys user.SysUser, secrets secret.FindGroupSecrets) nethost.Resolver {
	return func(name string) (nethost.Connection, error) {
		var found []secret.KeyValues
		for scr, err := range secrets(sys(), group.System) {
			if err != nil {
				return nethost.Connection{}, err
			}

			if kv, ok := scr.Credentials.(secret.KeyValues); ok && kv.Name == name {
				found = append(found, kv)
			}
		}

		switch len(found) {
		case 0:
			return nethost.Connection{}, fmt.Errorf("no connection %q is shared with the system group", name)
		case 1:
			return KeyValuesConnection(found[0])
		default:
			return nethost.Connection{}, fmt.Errorf("the name of the connection %q is ambiguous", name)
		}
	}
}

// KeyValuesConnection returns the connection of the secret. Without [secret.KeyBaseURL] the address is nil, so that the
// connection only offers its values. The decorator consumes the fixed keys, see [secret.IsFixedKey], and an empty value
// sets nothing, e.g. after the address has changed and a write-only value is gone. All other fields, which are not
// write-only, are the [nethost.Connection.Values].
func KeyValuesConnection(kv secret.KeyValues) (nethost.Connection, error) {
	conn := nethost.Connection{Values: map[string]string{}}
	headers := map[string]string{}
	query := map[string]string{}
	var bearer, basicUser, basicPassword string
	var secrets []string

	for field := range kv.Fields() {
		if !secret.IsFixedKey(field.Key) {
			if field.Protection == secret.WriteOnly {
				secrets = append(secrets, field.Value)
			} else {
				conn.Values[field.Key] = field.Value
			}
			continue
		}

		if field.Key == secret.KeyBaseURL {
			if field.Value != "" {
				base, err := parseBaseURL(field.Value)
				if err != nil {
					return nethost.Connection{}, err
				}
				conn.BaseURL = base
			}
			continue
		}

		if field.Protection != secret.Visible {
			secrets = append(secrets, field.Value)
		}

		switch field.Key {
		case secret.KeyHeaderBearerAuth:
			bearer = field.Value
			secrets = append(secrets, field.Value)
			continue
		case secret.KeyHeaderBasicAuthUser:
			basicUser = field.Value
			continue
		case secret.KeyHeaderBasicAuthPassword:
			basicPassword = field.Value
			secrets = append(secrets, field.Value)
			continue
		}

		if name, ok := secret.QueryForKey(field.Key); ok {
			if field.Value != "" {
				query[name] = field.Value
				if field.Protection != secret.Visible {
					// the transport sends this form, and a server may mirror the address
					secrets = append(secrets, url.QueryEscape(field.Value))
				}
			}
			continue
		}

		name, ok := secret.HeaderForKey(field.Key)
		if !ok {
			return nethost.Connection{}, fmt.Errorf("the key %q does not name a header of a connection", field.Key)
		}
		if field.Value != "" {
			headers[name] = field.Value
		}
	}

	if bearer != "" {
		headers["Authorization"] = "Bearer " + bearer
	}
	if basicUser != "" || basicPassword != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(basicUser + ":" + basicPassword))
		headers["Authorization"] = "Basic " + encoded
		if basicPassword != "" {
			secrets = append(secrets, encoded)
		}
	}

	conn.Decorate = setRequest(headers, query)
	conn.Secrets = scrubList(secrets)
	return conn, nil
}

// scrubList removes the empty and duplicate values. The longest comes first, because the host replaces them one after
// the other and a shorter value would otherwise leave the rest of a longer one, which contains it.
func scrubList(secrets []string) []string {
	secrets = slices.DeleteFunc(secrets, func(s string) bool { return s == "" })
	slices.SortFunc(secrets, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(b), len(a)), strings.Compare(a, b))
	})
	return slices.Compact(secrets)
}

func parseBaseURL(raw string) (*url.URL, error) {
	base, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("the base URL of the connection is invalid")
	}
	if base.Host == "" || base.User != nil || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, errors.New("the base URL of the connection must be http or https, have a host and no user")
	}
	return base, nil
}

// setRequest returns a decorator, which sets the headers and query parameters of each request, so that a program
// can not overwrite them. It returns nil, if there is nothing to set.
func setRequest(headers, query map[string]string) func(http.RoundTripper) http.RoundTripper {
	if len(headers) == 0 && len(query) == 0 {
		return nil
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(r *http.Request) (*http.Response, error) {
			r = r.Clone(r.Context())
			for name, value := range headers {
				r.Header.Set(name, value)
			}
			if len(query) > 0 {
				values := r.URL.Query()
				for name, value := range query {
					values.Set(name, value)
				}
				r.URL.RawQuery = values.Encode()
			}
			return next.RoundTrip(r)
		})
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
