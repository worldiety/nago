// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xerrors"
)

func NewCreate(mutex *sync.Mutex, bus events.Bus, repo Repository) Create {
	return func(subject auth.Subject, createOptions CreateOptions) (ID, error) {
		mutex.Lock()
		defer mutex.Unlock()

		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		id := createOptions.System.ID
		if id == "" {
			id = data.RandIdent[ID]()
		}

		ws := Workspace{
			ID:          id,
			Name:        createOptions.Name,
			Description: createOptions.Description,
			Platform:    createOptions.Platform,
			System:      createOptions.System.Valid,
		}

		if ws.Name == "" {
			return "", xerrors.WithFields("validation error", "Name", rstring.LabelValueMustNotBeEmpty.Get(subject))
		}

		switch ws.Platform {
		case MistralAI, OpenAI:
		// fine
		default:
			return "", fmt.Errorf("invalid platform: %s", ws.Platform)
		}

		if !ws.System {
			if optWs, err := repo.FindByID(ws.ID); err != nil || optWs.IsSome() {
				if err != nil {
					return "", err
				}

				if optWs.IsSome() {
					return "", fmt.Errorf("workspace %q already exists: %w", ws.ID, os.ErrNotExist)
				}
			}
		}

		if err := repo.Save(ws); err != nil {
			return "", err
		}

		return ws.ID, nil
	}
}
