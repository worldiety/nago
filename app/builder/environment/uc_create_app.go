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
	"sync"

	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewCreateApp(mutex *sync.Mutex, repoEnv Repository, repoApp app.Repository) CreateApp {
	return func(subject auth.Subject, env ID, opts CreateAppOptions) (app.ID, error) {
		mutex.Lock()
		defer mutex.Unlock()

		uEnv, err := getEnv(subject, repoEnv, env)
		if err != nil {
			return "", err
		}

		ap := app.App{
			ID:          data.RandIdent[app.ID](),
			Name:        opts.Name,
			Description: opts.Description,
		}

		if _, err := repoApp.FindByID(ap.ID); err != nil {
			if err != nil {
				return "", fmt.Errorf("failed to find app %v", ap.ID)
			}

			return "", fmt.Errorf("app already exists: %w", os.ErrNotExist)
		}

		if err := repoApp.Save(ap); err != nil {
			return "", fmt.Errorf("failed to save app %v: %w", ap.ID, err)
		}

		uEnv.Apps = append(uEnv.Apps, ap.ID)
		if err := repoEnv.Save(uEnv); err != nil {
			return "", fmt.Errorf("failed to save env %v: %w", env, err)
		}

		return ap.ID, nil
	}
}
