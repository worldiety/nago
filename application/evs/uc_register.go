// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"fmt"
	"reflect"
	"sync"

	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewRegister[Evt any](typeRegistry *concurrent.RWMap[reflect.Type, Discriminator], invTypeRegistry *concurrent.RWMap[Discriminator, reflect.Type]) Register[Evt] {
	var mutex sync.Mutex
	return func(t reflect.Type, discriminatorName Discriminator) error {
		mutex.Lock()
		defer mutex.Unlock()

		if err := discriminatorName.Validate(); err != nil {
			return err
		}

		if t == nil {
			return fmt.Errorf("type is required")
		}

		if !t.Implements(reflect.TypeFor[Evt]()) {
			return fmt.Errorf("type %v not implements type %v", t, reflect.TypeFor[Evt]())
		}

		if otherT, ok := invTypeRegistry.Get(discriminatorName); ok {
			if otherT == t {
				return nil // idempotent registrations
			}

			return fmt.Errorf("%s is already registered", discriminatorName)
		}

		if _, ok := typeRegistry.Get(t); ok {
			return fmt.Errorf("%v type is already registered", t)
		}

		typeRegistry.Put(t, discriminatorName)
		invTypeRegistry.Put(discriminatorName, t)
		return nil
	}
}
