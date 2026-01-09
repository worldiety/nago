// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/xtime"
)

func updateTimeSlice(inverseMutex *sync.Mutex, timeStore blob.Store, ts xtime.UnixMilliseconds, mut func(payload jsonInversePayload) jsonInversePayload) error {
	inverseMutex.Lock()
	defer inverseMutex.Unlock()

	tsKey, err := newTsKey(ts)
	if err != nil {
		return err
	}

	optBuf, err := blob.Get(timeStore, string(tsKey))
	if err != nil {
		return fmt.Errorf("error loading inverse lookup data %s: %w", tsKey, err)
	}

	var payload jsonInversePayload

	if optBuf.IsSome() {
		if err := json.Unmarshal(optBuf.Unwrap(), &payload); err != nil {
			return fmt.Errorf("error unmarshalling inverse payload: %w", err)
		}
	}

	// implementation note: this O(n) is limited to the amount of collisions for the same timestamp
	payload = mut(payload)
	if len(payload) == 0 {
		// purge the entire reverse key, if nothing left
		return blob.Delete(timeStore, string(tsKey))
	}

	slices.Sort(payload) // keep deterministic order, lower ids must be older
	buf := option.Must(json.Marshal(payload))
	if err := blob.Put(timeStore, string(tsKey), buf); err != nil {
		return fmt.Errorf("error updating inverse payload in store: %w", err)
	}

	return nil
}
