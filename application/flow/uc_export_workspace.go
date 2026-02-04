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
)

func NewExportWorkspace(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) ExportWorkspace {
	return func(subject auth.Subject, id WorkspaceID) ([]byte, error) {
		if err := subject.Audit(PermExportWorkspace); err != nil {
			return nil, err
		}

		var res ExportedWorkspace
		for evt, err := range handler.Replay(id) {
			if err != nil {
				return nil, err
			}

			res.Events = append(res.Events, evs.JsonEnvelope{
				Discriminator: evt.Discriminator,
				EventTime:     evt.EventTime,
				CreatedBy:     evt.CreatedBy,
				Metadata:      evt.Metadata,
				Data:          evt.Raw,
			})
		}

		return json.Marshal(res)
	}
}
