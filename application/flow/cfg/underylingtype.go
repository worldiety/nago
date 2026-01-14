// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import (
	"encoding/json"
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
)

type FieldInfo struct {
	ID             string
	Label          string
	SupportingText string
}

// TID is the underlying type identifier.
type TID string

// UnderlyingType defines a build-in data type which is the foundation for any [Field].
// These types must be declared at compile time and cannot be made dynamic.
// Marshal/Unmarshal expects one JSON Element (value, object, etc) so that it can get represented within
// an Object like {"JsonName":<RawMessage>}
type UnderlyingType[T any] interface {
	ID() TID
	Default() T
	Render(info FieldInfo, state *core.State[T]) core.View
	Validate(T) error // we also need a global validate
	Label() i18n.StrHnd
	Description() i18n.StrHnd

	Marshal(T) (json.RawMessage, error)
	Unmarshal(message json.RawMessage) (T, error)
}

func FieldType[T any](ut UnderlyingType[T]) Opt {
	return func(options *Options) {
		if options.underlyingTypes == nil {
			options.underlyingTypes = map[TID]UnderlyingType[any]{}
		}

		if _, ok := options.underlyingTypes[ut.ID()]; ok {
			panic(fmt.Errorf("duplicate field type: %s", ut.ID()))
		}

		if ut.ID() == "" {
			panic("field type must have an id")
		}

		options.underlyingTypes[ut.ID()] = anyUTypeWrapper[T]{ut}
	}
}

type anyUTypeWrapper[T any] struct {
	ut UnderlyingType[T]
}

func (a anyUTypeWrapper[T]) ID() TID {
	return a.ut.ID()
}

func (a anyUTypeWrapper[T]) Default() any {
	return a.ut.Default()
}

func (a anyUTypeWrapper[T]) Render(info FieldInfo, state *core.State[any]) core.View {
	// setup type-safe birectional binding
	cstate := core.DerivedState[T](state, ".any.T").Init(func() T {
		return a.ut.Default()
	}).Observe(func(newValue T) {
		state.Set(newValue)
		state.Notify()
	})

	state.Observe(func(newValue any) {
		if t, ok := newValue.(T); ok {
			cstate.Set(t)
		} else {
			cstate.Set(a.ut.Default())
		}

		cstate.Notify()
	})

	return a.ut.Render(info, cstate)
}

func (a anyUTypeWrapper[T]) Validate(t any) error {
	return a.ut.Validate(t.(T))
}

func (a anyUTypeWrapper[T]) Label() i18n.StrHnd {
	return a.ut.Label()
}

func (a anyUTypeWrapper[T]) Description() i18n.StrHnd {
	return a.ut.Description()
}

func (a anyUTypeWrapper[T]) Marshal(t any) (json.RawMessage, error) {
	return a.ut.Marshal(t.(T))
}

func (a anyUTypeWrapper[T]) Unmarshal(message json.RawMessage) (any, error) {
	return a.ut.Unmarshal(message)
}
