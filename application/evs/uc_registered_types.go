// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"iter"
	"reflect"
	"slices"

	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewRegisteredTypes[Evt any](registry *concurrent.RWMap[Discriminator, reflect.Type]) RegisteredTypes[Evt] {
	return func() iter.Seq[RegisteredType] {
		tmp := make([]Discriminator, 0, registry.Len())
		for k := range registry.All() {
			tmp = append(tmp, k)
		}

		slices.Sort(tmp)
		return func(yield func(RegisteredType) bool) {
			for _, s := range tmp {
				if v, ok := registry.Get(s); ok {
					if !yield(RegisteredType{
						Type:          v,
						Discriminator: s,
					}) {
						return
					}
				}
			}
		}
	}
}
