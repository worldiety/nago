// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
)

func NewMatch(repo Repository) Match {
	return func(subject auth.Subject, typ reflect.Type, opts MatchOptions) (option.Opt[Credentials], error) {
		var best *Secret
		for secret, err := range repo.All() {
			if err != nil {
				return option.Opt[Credentials]{}, err
			}

			if reflect.TypeOf(secret.Credentials) != typ {
				continue
			}

			if opts.Group != "" {
				if !slices.Contains(secret.Groups, opts.Group) {
					return option.None[Credentials](), nil
				}
			}

			if best == nil {
				best = &secret
			} else {
				if opts.Hint != "" && (string(secret.ID) == opts.Hint || secret.Credentials.GetName() == opts.Hint) {
					best = &secret
				}
			}
		}

		if best == nil {
			if opts.Expect {
				if opts.Group != "" {
					return option.None[Credentials](), fmt.Errorf("no %q credential found, check the secrets and if it shared with group %q", typ, opts.Group)
				}

				return option.None[Credentials](), fmt.Errorf("no %q credential found, check the secrets", typ)
			}

			return option.None[Credentials](), nil
		}

		return option.Some(best.Credentials), nil
	}
}
