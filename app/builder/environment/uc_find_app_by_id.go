// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package environment

import (
	"fmt"
	"os"
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/auth"
)

func NewFindAppByID(repoEnv Repository, repoApp app.Repository) FindAppByID {
	return func(subject auth.Subject, env ID, id app.ID) (option.Opt[app.App], error) {
		var zero option.Opt[app.App]
		uEnv, err := getEnv(subject, repoEnv, env)
		if err != nil {
			return zero, err
		}

		if !slices.Contains(uEnv.Apps, id) {
			return zero, fmt.Errorf("app %s is not owned by environment %s: %w", id, env, os.ErrNotExist)
		}

		optApp, err := repoApp.FindByID(id)
		if err != nil {
			return zero, fmt.Errorf("failed to find app by id: %w", err)
		}

		return optApp, nil
	}
}
