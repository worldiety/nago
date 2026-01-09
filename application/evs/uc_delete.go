// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewDelete[Evt any](perms Permissions, inverseMutex *sync.Mutex, load Load[Evt], eventStore blob.Store, timeStore blob.Store, gOpts Options[Evt]) Delete[Evt] {
	return func(subject auth.Subject, id SeqID) error {
		if err := subject.Audit(perms.Delete); err != nil {
			return err
		}

		optEvt, err := load(user.SU(), id)
		if err != nil {
			return fmt.Errorf("failed to load event %v: %w", id, err)
		}

		if optEvt.IsNone() {
			return nil
		}

		evt := optEvt.Unwrap()

		seq, err := NewSeqKey(id)
		if err != nil {
			return err
		}

		ctx := context.Background()
		if err := eventStore.Delete(ctx, string(seq)); err != nil {
			return fmt.Errorf("failed to delete event: %w", err)
		}

		// cleanup the inverse
		if err := updateTimeSlice(inverseMutex, timeStore, evt.EventTime, func(payload jsonInversePayload) jsonInversePayload {
			return slices.DeleteFunc(payload, func(sid SeqID) bool {
				return sid == id
			})
		}); err != nil {
			return err
		}

		// purge from indices
		for _, idxer := range gOpts.Indexer {
			if err := idxer.Remove(evt); err != nil {
				return fmt.Errorf("failed to remove %v from index (%T): %w", id, idxer, err)
			}
		}

		return nil
	}
}
