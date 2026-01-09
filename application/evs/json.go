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

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

type jsonEnvelope struct {
	Discriminator Discriminator          `json:"t,omitempty"`
	EventTime     xtime.UnixMilliseconds `json:"ts,omitempty"`
	CreatedBy     user.ID                `json:"createdBy,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
	Data          json.RawMessage        `json:"data,omitempty"`
}

func (e jsonEnvelope) decodeData(registry *concurrent.RWMap[Discriminator, reflect.Type]) (any, error) {
	rtype, ok := registry.Get(e.Discriminator)
	if !ok {
		return nil, fmt.Errorf("unknown type: %s", e.Discriminator)
	}

	rval := reflect.New(rtype)
	if err := json.Unmarshal(e.Data, rval.Interface()); err != nil {
		return nil, err
	}

	return rval.Elem().Interface(), nil
}

type jsonInversePayload []SeqID
