// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"context"
	"iter"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/xiter"
)

func NewFindInstanceEvents(events blob.Store) FindInstanceEvents {
	return func(subject user.Subject, id Instance) iter.Seq2[EventKey, error] {
		if err := subject.AuditResource(events.Name(), string(id), PermFindInstanceEvents); err != nil {
			return xiter.WithError[EventKey](err)
		}

		return func(yield func(EventKey, error) bool) {
			for evtId, err := range events.List(context.Background(), blob.ListOptions{Prefix: string(id)}) {
				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				if !yield(EventKey(evtId), nil) {
					return
				}
			}
		}
	}
}
