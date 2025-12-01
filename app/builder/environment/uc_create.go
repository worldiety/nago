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

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewCreate(mutex *sync.Mutex, repo Repository) Create {
	return func(subject auth.Subject, opts CreateOptions) (ID, error) {
		// security note: every authenticated user is allowed to create an environment
		if !subject.Valid() {
			return "", user.InvalidSubjectErr
		}

		mutex.Lock()
		defer mutex.Unlock()

		env := Environment{
			ID:          data.RandIdent[ID](),
			Name:        opts.Name,
			Description: opts.Description,
			Owner:       []user.ID{subject.ID()},
		}

		if optEnv, err := repo.FindByID(env.ID); err != nil || optEnv.IsSome() {
			if err != nil {
				return "", fmt.Errorf("cannot check existing environment: %w", err)
			}

			return "", fmt.Errorf("environment already exists: %s: %w", env.ID, os.ErrExist)
		}

		if err := repo.Save(env); err != nil {
			return "", fmt.Errorf("cannot save environment: %w", err)
		}

		return env.ID, nil
	}
}
