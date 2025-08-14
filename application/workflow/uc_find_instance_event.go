// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

func NewFindInstanceEvent(events blob.Store) FindInstanceEvent {
	return func(subject auth.Subject, key EventKey) (option.Opt[Event], error) {
		if err := subject.Audit(PermFindInstanceEvents); err != nil {
			return option.Opt[Event]{}, err
		}

		optEvt, err := json.Get[persistedEventData](events, string(key))
		if err != nil {
			return option.Opt[Event]{}, err
		}

		if optEvt.IsNone() {
			return option.Opt[Event]{}, nil
		}

		evt := optEvt.Unwrap()
		inst, seq, err := key.Split()
		if err != nil {
			return option.Opt[Event]{}, err
		}

		return option.Some(Event{
			Payload:  evt.Payload.Value,
			Instance: inst,
			SavedAt:  evt.SavedAt,
			SeqNo:    seq,
			ID:       key,
		}), nil
	}
}
