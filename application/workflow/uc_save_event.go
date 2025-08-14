// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xjson"
	"go.wdy.de/nago/pkg/xtime"
)

type EventKey string

func NewEventKey(id Instance, seq int64) EventKey {
	return EventKey(string(id) + "/" + strconv.FormatInt(seq, 10))
}

func (k EventKey) Split() (Instance, int64, error) {
	token := strings.Split(string(k), "/")
	if len(token) != 2 {
		return "", 0, fmt.Errorf("invalid event key")
	}

	seq, err := strconv.ParseInt(token[1], 10, 64)
	if err != nil {
		return "", 0, err
	}

	return Instance(token[0]), seq, nil
}

func NewSaveEvent(events blob.Store) SaveEvent {
	var lastEvtSequenceIds concurrent.RWMap[Instance, int64]
	var mutex sync.Mutex

	return func(subject user.Subject, id Instance, evt any) error {
		mutex.Lock()
		defer mutex.Unlock()

		if env, ok := evt.(InstanceEventEnvelope); ok {
			evt = env
		}

		lastSeqNo, ok := lastEvtSequenceIds.Get(id)
		if !ok {
			for key, err := range events.List(context.Background(), blob.ListOptions{Prefix: string(id)}) {
				if err != nil {
					return fmt.Errorf("cannot list events: %w", err)
				}

				_, seq, err := EventKey(key).Split()
				if err != nil {
					return fmt.Errorf("cannot split key '%s': %w", key, err)
				}

				if lastSeqNo < seq {
					lastSeqNo = seq
					lastEvtSequenceIds.Put(id, seq)
				}
			}
		}

		lastSeqNo++
		lastEvtSequenceIds.Put(id, lastSeqNo)

		key := NewEventKey(id, lastSeqNo)
		pevt := persistedEventData{
			SavedAt: xtime.UnixMilliseconds(time.Now().UnixMilli()),
			Payload: xjson.NewAdjacentEnvelope(evt),
		}

		buf, err := json.Marshal(pevt)
		if err != nil {
			return fmt.Errorf("cannot marshal event: %w", err)
		}

		if err := blob.Put(events, string(key), buf); err != nil {
			return fmt.Errorf("cannot put event: %w", err)
		}

		return nil
	}
}
