// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"encoding/json"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
)

func NewImportWorkspace(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) ImportWorkspace {
	return func(subject auth.Subject, data []byte) error {
		if err := subject.Audit(PermImportWorkspace); err != nil {
			return err
		}

		var exp ExportedWorkspace
		if err := json.Unmarshal(data, &exp); err != nil {
			return err
		}

		return handler.Restore(subject.Context(), xslices.ValuesWithError(exp.Events, nil))
	}
}
