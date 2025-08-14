// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

func NewGetStatus(eventStore blob.Store) GetStatus {
	return func(subject auth.Subject, id Instance) (option.Opt[Status], error) {
		if err := subject.AuditResource(eventStore.Name(), string(id), PermGetStatus); err != nil {
			return option.Opt[Status]{}, err
		}

		var status Status
		status.ID = id

		for evtId, err := range eventStore.List(context.Background(), blob.ListOptions{Prefix: string(id)}) {
			if err != nil {
				return option.Opt[Status]{}, fmt.Errorf("cannot list events for instance: %w", err)
			}

			optEvt, err := json.Get[persistedEventData](eventStore, evtId)
			if err != nil {
				return option.Opt[Status]{}, fmt.Errorf("cannot get event data: %s: %w", evtId, err)
			}

			if optEvt.IsNone() {
				// concurrent stale reference
				continue
			}

			evt := optEvt.Unwrap()
			switch evt := evt.Payload.Value.(type) {
			case InstanceStopped:
				status.State = InstanceDone
				status.StoppedAt = optEvt.Unwrap().SavedAt.Time(time.Local)
			case InstanceCreated:
				status.State = InstanceRunning
				status.StartedAt = optEvt.Unwrap().SavedAt.Time(time.Local)
			case ActionFailed:
				status.Error = Error{
					Valid:   true,
					Message: evt.Error,
					Type:    evt.ErrorType,
				}
			}
		}

		return option.Some(status), nil
	}
}
