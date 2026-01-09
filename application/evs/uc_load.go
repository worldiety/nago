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
	"reflect"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewLoad[Evt any](perms Permissions, eventStore blob.Store, registry *concurrent.RWMap[Discriminator, reflect.Type]) Load[Evt] {
	return func(subject auth.Subject, id SeqID) (option.Opt[Envelope[Evt]], error) {
		var zero option.Opt[Envelope[Evt]]
		if !subject.HasPermission(perms.Load) && !subject.HasPermission(perms.ReadAll) {
			return zero, subject.Audit(perms.Load)
		}

		key, err := NewSeqKey(id)
		if err != nil {
			return zero, err
		}

		optBuf, err := blob.Get(eventStore, string(key))
		if err != nil {
			return zero, err
		}

		if optBuf.IsNone() {
			return zero, nil
		}

		var jsonEnv jsonEnvelope
		if err := json.Unmarshal(optBuf.Unwrap(), &jsonEnv); err != nil {
			return zero, err
		}

		payload, err := jsonEnv.decodeData(registry)
		if err != nil {
			return zero, err
		}

		payloadEvt, ok := payload.(Evt)
		if !ok {
			regType, _ := registry.Get(jsonEnv.Discriminator)
			return zero, fmt.Errorf("dynamic type mismatch in envelope: registered type %s=%v is not convertible into %T: you probably renamed or refactored something in an incompatible way", jsonEnv.Discriminator, regType, jsonEnv.Discriminator)
		}

		return option.Some(Envelope[Evt]{
			Sequence:      id,
			Key:           key,
			Discriminator: jsonEnv.Discriminator,
			EventTime:     jsonEnv.EventTime,
			CreatedBy:     jsonEnv.CreatedBy,
			Metadata:      jsonEnv.Metadata,
			Data:          payloadEvt,
			Raw:           optBuf.Unwrap(),
		}), nil
	}
}
