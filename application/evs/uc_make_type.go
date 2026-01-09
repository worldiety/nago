// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"fmt"
	"os"
	"reflect"

	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewMakeType[Evt any](invTypeRegistry *concurrent.RWMap[Discriminator, reflect.Type]) MakeType[Evt] {
	return func(discriminator Discriminator) (Evt, error) {
		var zero Evt
		t, ok := invTypeRegistry.Get(discriminator)
		if !ok {
			return zero, fmt.Errorf("unknown type: %s: %w", discriminator, os.ErrNotExist)
		}

		rval := reflect.New(t)
		evt, ok := rval.Elem().Interface().(Evt)
		if !ok {
			return zero, fmt.Errorf("dynamic type mismatch: registered type %s=%v is not convertible into %T: you probably renamed or refactored something in an incompatible way", discriminator, t, zero)
		}

		return evt, nil
	}
}
