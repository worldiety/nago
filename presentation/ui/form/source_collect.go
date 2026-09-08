// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"log/slog"

	"go.wdy.de/nago/application/user"
)

// collectSourceEntities walks a [Source] the way the picker renderers need it: list the identities, then
// resolve each one into the entity that carries the display value.
//
// An identity that does not resolve is skipped rather than fatal. That happens for perfectly ordinary
// reasons - the entry was removed between the listing and the lookup, or the source lists more than it can
// resolve - and it used to take the whole page down: the renderers called Unwrap on the optional without
// checking it, which panics with "unwrapped invalid option" instead of producing a picker with one option
// less.
func collectSourceEntities(source Source, subject user.Subject) ([]Entity, error) {
	var values []Entity

	for id, err := range source.FindAll(subject) {
		if err != nil {
			return nil, err
		}

		optE, err := source.FindByID(subject, id)
		if err != nil {
			return nil, err
		}

		if optE.IsNone() {
			// Logged, not swallowed silently: a source that regularly lists what it cannot resolve is a bug
			// worth finding, just not one worth crashing over.
			slog.Error("form.Auto source listed an id it cannot resolve", "id", id)
			continue
		}

		values = append(values, optE.Unwrap())
	}

	return values, nil
}
