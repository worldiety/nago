// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

func NewStore[Evt any](perms Permissions, inverseMutex *sync.Mutex, typeRegistry *concurrent.RWMap[reflect.Type, Discriminator], eventStore blob.Store, timeStore blob.Store, gOpts Options[Evt]) Store[Evt] {
	var lastId atomic.Int64
	var once sync.Once

	nextID := func() (SeqID, error) {
		var initErr error
		once.Do(func() {
			// even though this looks like an O(n) loop, we inverted the iteration order and just pick
			// the highest sequence number below
			for sid, err := range eventStore.List(context.Background(), blob.ListOptions{
				Reverse: true,
			}) {

				if err != nil {
					initErr = fmt.Errorf("error listing events: %w", err)
					return
				}

				key := SeqKey(sid)
				n, err := key.Parse()
				if err != nil {
					initErr = fmt.Errorf("error parsing event key %s: %w", sid, err)
					return
				}

				lastId.Store(int64(n))
				break
			}
		})

		if initErr != nil {
			return 0, initErr
		}

		return SeqID(lastId.Add(1)), nil
	}

	return func(subject auth.Subject, evt Evt, opts StoreOptions) (Envelope[Evt], error) {
		var zero Envelope[Evt]
		if err := subject.Audit(perms.Store); err != nil {
			return zero, err
		}

		discriminator, ok := typeRegistry.Get(reflect.TypeOf(evt))
		if !ok {
			baseType := reflect.TypeFor[Evt]().String()
			// cfgevs.Schema[flow.PrimaryKeySelected, flow.WorkspaceEvent]("PrimaryKeySelected")
			return zero, fmt.Errorf("type %T not found in type registry. Use cfgevs.Schema[%T, %s](\"My Alias\") to declare it", evt, evt, baseType)
		}

		payloadBuf, err := json.Marshal(evt)
		if err != nil {
			return zero, fmt.Errorf("event %T cannot be marshalled: %w", evt, err)
		}

		mySeqId, err := nextID()
		if err != nil {
			return zero, err
		}

		key, err := NewSeqKey(mySeqId)
		if err != nil {
			return zero, err
		}

		if opts.CreatedBy == "" {
			opts.CreatedBy = subject.ID()
		}

		if opts.EventTime == 0 {
			opts.EventTime = xtime.Now()
		}

		env := JsonEnvelope{
			Discriminator: discriminator,
			EventTime:     opts.EventTime,
			CreatedBy:     opts.CreatedBy,
			Metadata:      opts.Metadata,
			Data:          payloadBuf,
		}

		buf, err := json.Marshal(env)
		if err != nil {
			return zero, fmt.Errorf("error marshalling envelope: %w", err)
		}

		if err := blob.Put(eventStore, string(key), buf); err != nil {
			return zero, fmt.Errorf("error storing envelope in store: %w", err)
		}

		// put the inverse
		if err := updateTimeSlice(inverseMutex, timeStore, env.EventTime, func(payload jsonInversePayload) jsonInversePayload {
			payload = append(payload, mySeqId)
			return payload
		}); err != nil {
			return zero, err
		}

		// update composite indicies
		e := Envelope[Evt]{
			Sequence:      mySeqId,
			Key:           key,
			Discriminator: env.Discriminator,
			EventTime:     env.EventTime,
			CreatedBy:     env.CreatedBy,
			Metadata:      env.Metadata,
			Data:          evt,
			Raw:           buf,
		}

		for _, idxer := range gOpts.Indexer {
			if err := idxer.Insert(e); err != nil {
				// we are essentially screwed: if the write fails, usually due to disk full, we cannot even insert any
				// deletes for the prior inserts
				return e, err
			}
		}

		return e, nil
	}
}
