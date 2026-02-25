// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"sync/atomic"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xerrors"
	"go.wdy.de/nago/presentation/core"
)

// Renderer is a stateless function that renders a field. If it returns nil, it is not responsible and
// continues with the next field. The first renderer which can render the field wins thus registration order is
// significant.
type Renderer func(ctx FieldContext) core.View

var (
	// Renderers contain the default renderers for auto form. Do not change this iterator, instead
	// consider passing your own using the AutoOptions.
	Renderers = slices.Values([]Renderer{
		RenderHLine,
		RenderPlaintext,
		RenderSourceSliceString,
		RenderSourceString,
		RenderXTimeDate,
		RenderColor,
		RenderImage,
		RenderSliceText,
		RenderText,
		RenderFloat,
		RenderDuration,
		RenderBool,
		RenderInt,
		RenderXTimeTimeFrame,
	})
)

// FieldContext contains all information about a field which has been parsed from a struct field
// improve render performance by caching the parsed values.
type FieldContext struct {
	field          reflect.StructField
	disabled       bool
	label          string
	supportingText string
	value          string
	values         []string
	id             string
	source         Source
	wnd            core.Window
	ctx            context.Context
	state          *core.State[any]
	errorText      string
}

func NewFieldContext[T any](wnd core.Window, opts AutoOptions, state *core.State[T], field reflect.StructField) FieldContext {
	disabled := false
	if flag, ok := field.Tag.Lookup("disabled"); ok && flag == "true" {
		disabled = true
	}

	if opts.ViewOnly {
		disabled = true
	}

	if opts.Window != nil {
		wnd = opts.Window
	}

	ctx := wnd.Context()
	if opts.Context != nil {
		ctx = opts.Context
	}

	label := field.Name
	if name, ok := field.Tag.Lookup("label"); ok {
		label = name
	}

	// try to translate
	label = wnd.Bundle().Resolve(label)
	supportingText := wnd.Bundle().Resolve(field.Tag.Get("supportingText"))

	var values []string
	if array, ok := field.Tag.Lookup("values"); ok {
		err := json.Unmarshal([]byte(array), &values)
		if err != nil {
			slog.Error("cannot parse values from struct field", "type", fmt.Sprintf("%T", state.Get()), "field", field.Name, "literal", array, "err", err.Error())
		}
	}

	value := field.Tag.Get("value")

	var src Source
	if sourceName := field.Tag.Get("source"); sourceName != "" {
		source, ok := core.FromContext[Source](ctx, sourceName)
		if ok {
			src = source
		} else {
			ucAny, ok := core.FromContext[UseCaseListAny](ctx, sourceName)
			if ok {
				src = wrapUseCaseListAny(ucAny)
			}
		}
	}

	if src == nil && len(values) > 0 {
		src = wrapStrSlice(values)
	}

	var errorText string

	if err, ok := errors.AsType[xerrors.ErrorWithFields](opts.Errors); ok {
		errorText = err.Fields[field.Name]
	}

	anyState := core.DerivedState[any](state, "-any").Init(func() any {
		return state.Get()
	}).Observe(func(newValue any) {
		state.Set(newValue.(T))
		state.Notify()
	})

	return FieldContext{
		wnd:            wnd,
		disabled:       disabled,
		ctx:            ctx,
		label:          label,
		supportingText: supportingText,
		value:          value,
		values:         values,
		id:             field.Tag.Get("id"),
		source:         src,
		errorText:      errorText,
		state:          anyState,
		field:          field,
	}
}

// Disabled only returns true, if the field has been tagged with disabled:"true"
func (c FieldContext) Disabled() bool {
	return c.disabled
}

// Label either returns the field name or the value of the label tag.
func (c FieldContext) Label() string {
	return c.label
}

// Value returns the the value literal of the value tag.
func (c FieldContext) Value() string {
	return c.value
}

// Values returns the values of the values tag. It must be json-array encoded like values:"[\"a\", \"b\"]".
func (c FieldContext) Values() []string {
	return c.values
}

// ID returns the literal of the id tag and is used as the component id.
func (c FieldContext) ID() string {
	return c.id
}

// Source returns whatever the source tag contains and whatever can be resolved. This is nil if no source is specified
// or the source cannot be resolved. The following source are automatically converted into the source interface:
//   - Value (field tag value)
//   - Values (field tag values)
//   - UseCaseListAny (from context via field tag source)
//   - Source (from context via field tag source)
func (c FieldContext) Source() Source {
	return c.source
}

// Window returns either the window from the options or the current window of rendering.
func (c FieldContext) Window() core.Window {
	return c.wnd
}

// Context either returns the context from the current window or the context from the options.
func (c FieldContext) Context() context.Context {
	return c.ctx
}

// SupportingText returns the literal of the supportingText tag.
func (c FieldContext) SupportingText() string {
	return c.supportingText
}

func (c FieldContext) Field() reflect.StructField {
	return c.field
}

func (c FieldContext) State() *core.State[any] {
	return c.state
}

func (c FieldContext) ErrorText() string {
	return c.errorText
}

func (c FieldContext) Subject() user.Subject {
	return c.wnd.Subject()
}

func (c FieldContext) IsSlice(kind reflect.Kind) bool {
	if c.field.Type.Kind() != reflect.Slice {
		return false
	}

	return c.field.Type.Elem().Kind() == kind
}

func (c FieldContext) IsType(t reflect.Type) bool {
	return c.field.Type == t
}

// SetValue tries to coerce the given value into the current field and updates and notifies the entire state.
// The value is automatically converted to the field's type if possible (e.g. string to a named string type).
// When used wrong, this may bail out with a panic.
func (c FieldContext) SetValue(value any) {
	dst := c.state.Get()

	vDst := reflect.ValueOf(dst)

	for vDst.Kind() == reflect.Ptr || vDst.Kind() == reflect.Interface {
		vDst = vDst.Elem()
	}

	cpy := reflect.New(vDst.Type()).Elem()
	cpy.Set(vDst)

	fieldVal := cpy.FieldByIndex(c.field.Index)
	rVal := reflect.ValueOf(value)
	if rVal.Type() != fieldVal.Type() && rVal.Type().ConvertibleTo(fieldVal.Type()) {
		rVal = rVal.Convert(fieldVal.Type())
	}
	fieldVal.Set(rVal)

	c.state.Set(cpy.Interface())
	c.state.Notify()
}

// Source is a more performant alternative to [UseCaseListAny] which allows partial loading of the entire entity set.
// See also [NewSource].
type Source interface {
	FindAll(subject user.Subject) iter.Seq2[string, error]
	FindByID(subject user.Subject, id string) (option.Opt[Entity], error)
}

type funcAnySource struct {
	findAll  func(subject user.Subject) iter.Seq2[string, error]
	findByID func(subject user.Subject, id string) (option.Opt[Entity], error)
}

func (f funcAnySource) FindAll(subject user.Subject) iter.Seq2[string, error] {
	return f.findAll(subject)
}

func (f funcAnySource) FindByID(subject user.Subject, id string) (option.Opt[Entity], error) {
	return f.findByID(subject, id)
}

// NewSource creates a new Source from the given generic functions.
func NewSource[T data.Aggregate[ID], ID ~string](findAll func(subject user.Subject) iter.Seq2[ID, error], findByID func(subject user.Subject, id ID) (option.Opt[T], error)) Source {
	return funcAnySource{
		findAll: func(subject user.Subject) iter.Seq2[string, error] {
			return func(yield func(string, error) bool) {
				for id, err := range findAll(subject) {
					if !yield(string(id), err) {
						return
					}
				}
			}
		},
		findByID: func(subject user.Subject, id string) (option.Opt[Entity], error) {
			optE, err := findByID(subject, ID(id))
			if err != nil {
				return option.None[Entity](), err
			}

			if optE.IsNone() {
				return option.None[Entity](), nil
			}

			e := optE.Unwrap()
			return option.Some(Entity{
				ID:    string(e.Identity()),
				Value: e,
			}), nil
		},
	}
}

// wrapUseCaseListAny requests all entities for each findAll request and caches the result so that
// findByID calls can be answered in O(1). This allows us to just implement against Source and be still
// reasonably performant.
func wrapUseCaseListAny(uc UseCaseListAny) Source {
	var tmp atomic.Pointer[map[string]AnyEntity]
	var tmpOrder atomic.Pointer[[]string]

	return funcAnySource{
		findAll: func(subject user.Subject) iter.Seq2[string, error] {
			m := map[string]AnyEntity{}
			var order []string
			var e error
			for v, err := range uc(subject) {
				if err != nil {
					e = err
					break
				}

				m[v.Identity()] = v
				order = append(order, v.Identity())
			}

			tmp.Store(&m)
			tmpOrder.Store(&order)

			return func(yield func(string, error) bool) {
				if e != nil {
					yield("", e)
					return
				}

				for _, s := range order {
					if !yield(s, nil) {
						return
					}
				}
			}
		},
		findByID: func(subject user.Subject, id string) (option.Opt[Entity], error) {
			mPtr := tmp.Load()
			if mPtr == nil {
				return option.None[Entity](), nil
			}

			m := *mPtr
			v, ok := m[id]
			if !ok {
				return option.None[Entity](), nil
			}

			return option.Some(Entity{
				ID:    v.id,
				Value: v.aggregate,
			}), nil
		},
	}
}

// wrapSliceIndex artificially wraps a slice into a Source and models the index as string key.
func wrapStrSlice(slice []string) Source {
	return funcAnySource{
		findAll: func(subject user.Subject) iter.Seq2[string, error] {
			return func(yield func(string, error) bool) {
				for _, s := range slice {
					if !yield(s, nil) {
						return
					}
				}
			}
		},
		findByID: func(subject user.Subject, id string) (option.Opt[Entity], error) {
			return option.Some(Entity{
				ID:    id,
				Value: id,
			}), nil
		},
	}
}

// wrapSliceIndex artificially wraps a slice into a Source and models the index as string key.
func wrapSliceIndex[T any](slice []T) Source {
	return funcAnySource{
		findAll: func(subject user.Subject) iter.Seq2[string, error] {
			return func(yield func(string, error) bool) {
				for idx := range slice {
					if !yield(strconv.Itoa(idx), nil) {
						return
					}
				}
			}
		},
		findByID: func(subject user.Subject, id string) (option.Opt[Entity], error) {
			idx, err := strconv.Atoi(id)
			if err != nil {
				return option.None[Entity](), err
			}

			if idx < 0 || idx >= len(slice) {
				return option.None[Entity](), nil
			}

			return option.Some(Entity{
				ID:    id,
				Value: slice[idx],
			}), nil
		},
	}
}
