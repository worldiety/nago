// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package environment

import (
	"fmt"
	"iter"
	"os"

	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/auth"
)

func NewReplay(evtRepo EventRepository, findAppByID FindAppByID) Replay {
	return func(subject auth.Subject, env ID, app app.ID) iter.Seq2[EventBox, error] {
		return func(yield func(EventBox, error) bool) {
			optApp, err := findAppByID(subject, env, app)
			if err != nil {
				yield(EventBox{}, err)
				return
			}

			if optApp.IsNone() {
				yield(EventBox{}, fmt.Errorf("app is none: %w", os.ErrNotExist))
				return
			}

			for box, err := range evtRepo.FindAllByPrefix(EID(app)) {
				if err != nil {
					if !yield(box, err) {
						return
					}

					continue
				}

				if !yield(box, nil) {
					return
				}
			}
		}
	}
}
