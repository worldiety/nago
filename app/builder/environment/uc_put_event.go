// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package environment

import (
	"fmt"
	"os"

	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/auth"
)

func NewPutEvent(evtRepo EventRepository, findAppByID FindAppByID) PutEvent {
	return func(subject auth.Subject, env ID, app app.ID, event Event) error {
		optApp, err := findAppByID(subject, env, app)
		if err != nil {
			return err
		}

		if optApp.IsNone() {
			return fmt.Errorf("app %s not found: %w", app, os.ErrNotExist)
		}

		box := NewEventBox(subject, app, event)
		return evtRepo.Save(box)
	}
}
