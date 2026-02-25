// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"context"
	"errors"
	"iter"
	"reflect"

	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xerrors"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

type AutoOptions struct {
	SectionPadding std.Option[ui.Padding]
	ViewOnly       bool
	IgnoreFields   []string

	// Specific Window to use, if nil uses the default window at render time.
	Window core.Window

	// Context is used to resolve the data sources. If Context is nil, the Window context is used to resolve the sources.
	Context context.Context

	// Errors can be either a map[string]string or a struct {Field string} to resolve error messages for
	// specific fields. In both cases, the key respective the Field names must match the original introspected
	// field names of the model. See also [ErrorWithFields]. Any other error will be appended as alert.BannerError.
	// See also [xerrors.WithFields] to create compatible wrapper error.
	Errors error

	// Renderers can be used to customize the rendering of specific field types. If empty, all default renderers are
	// used from [Renderers].
	Renderers iter.Seq[Renderer]
}

func (o AutoOptions) context() context.Context {
	if o.Context != nil {
		return o.Context
	}

	if o.Window != nil {
		return o.Window.Context()
	}

	return context.Background()
}

// TAuto is a composite component (Auto Form).
// This component renders a form for type T driven by reflection,
// bound to a state and configurable via AutoOptions.
type TAuto[T any] struct {
	opts  AutoOptions    // options controlling form generation and behavior
	state *core.State[T] // bound state holding the form model

	padding            ui.Padding // layout padding
	frame              ui.Frame   // frame defining size and layout
	border             ui.Border  // border styling
	accessibilityLabel string     // accessibility label for screen readers
	invisible          bool       // whether the form is hidden
	cardPadding        ui.Padding
}

// Auto is similar to [crud.AutoBinding], however it does much less and just creates a form using
// reflection from the given type. It does not require or understand entities and identities.
// Also note, that the concrete type is inspected at runtime and not the given template T, which
// is only needed for your convenience and to satisfy any concrete state type. Internally, everything gets evaluated
// as [any]. T maybe also be an interface, thus ensure, that the state contains not a nil interface.
//
// The current default implementation only supports:
//   - string fields
//   - integer fields (literally)
//   - string slices
//   - bool fields
//   - float fields
//
// Other features, which are supported by [crud.Auto] are not (yet) supported.
//
// Supported field tags:
//   - visible:"true"|"false" defaults to true
//   - section:"some text" defaults to zero
//   - label:"string literal"|"i18n key" defaults to Field name
//   - source:"source id" defaults to zero, only applicable to fields with underlying type string or []string. The
//     source must be provided using [Configuration.AddContextValue] as type [AnyUseCaseList].
//   - lines:"integer" only applicable to fields with underlying type string or []string and defaults to zero which
//     renders as a single line. 1 also renders as single line, but uses a multiline input element. Defaults to zero
//     for string types and to 5 for []string types.
//   - value:"string literal"|"bool literal"|"number literal" only applicable for fields with the according underlying
//     type. Defaults to the zero value of the underlying type.
//   - dialogOptions:"large|larger|xlarge|xxlarge" is only supported for source picker.
//
// The actual support may vary and depends on [AutoOptions.Renderers].
func Auto[T any](opts AutoOptions, state *core.State[T]) TAuto[T] {
	return TAuto[T]{
		opts:        opts,
		state:       state,
		cardPadding: ui.Padding{Right: ui.L40, Left: ui.L40, Bottom: ui.L40, Top: ""},
	}
}

// Padding sets the padding of the auto form.
func (t TAuto[T]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TAuto[T]) CardPadding(padding ui.Padding) TAuto[T] {
	t.cardPadding = padding
	return t
}

// WithFrame updates the frame of the auto form using a transformation function.
func (t TAuto[T]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

// Frame sets the frame of the auto form directly.
func (t TAuto[T]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TAuto[T]) FullWidth() TAuto[T] {
	t.frame.Width = ui.Full
	return t
}

// Border sets the border styling of the auto form.
func (t TAuto[T]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

// Visible toggles the visibility of the auto form.
func (t TAuto[T]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

// AccessibilityLabel sets the accessibility label for the auto form.
func (t TAuto[T]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}

func (t TAuto[T]) Render(ctx core.RenderContext) core.RenderNode {
	if t.opts.Window == nil {
		t.opts.Window = ctx.Window()
	}

	// TODO can we unify this with the crud package, but it is so different under the hood and equal at the same time?
	value := any(t.state.Get())
	if value == nil {
		var zero T
		value = zero
	}

	if value == nil {
		return ui.VStack(alert.Banner("implementation error", "no type information available for [form.Auto]")).Render(ctx)
	}

	var err error
	if _, ok := errors.AsType[xerrors.ErrorWithFields](t.opts.Errors); !ok {
		err = t.opts.Errors
	}

	var rootViews xslices.Builder[core.View]
	structType := reflect.TypeOf(value)
	rIt := t.opts.Renderers
	if rIt == nil {
		rIt = Renderers
	}

	for _, group := range LocalizeGroups(ctx.Window().Bundle(), GroupsOf(structType, t.opts.IgnoreFields...)) {
		var fieldsBuilder xslices.Builder[core.View]
	NextField:
		for _, field := range group.Fields {
			fctx := NewFieldContext[T](ctx.Window(), t.opts, t.state, field)

			for renderer := range rIt {
				if v := renderer(fctx); v != nil {
					fieldsBuilder.Append(v)
					continue NextField
				}
			}
		}

		fields := fieldsBuilder.Collect()
		if group.Name == "" {
			rootViews.Append(fields...)
		} else {
			card := cardlayout.Card(group.Name).Padding(t.cardPadding).Body(ui.VStack(fields...).Gap(ui.L16).FullWidth()).Frame(ui.Frame{}.FullWidth())
			if t.opts.SectionPadding.IsSome() {
				card = card.Padding(t.opts.SectionPadding.Unwrap())
			}
			rootViews.Append(card)
		}
	}

	if err != nil {
		rootViews.Append(alert.BannerError(err))
	}

	return ui.VStack(rootViews.Collect()...).Gap(ui.L16).FullWidth().Render(ctx)
}

func getDialogOptions(field reflect.StructField) []alert.Option {
	var dlgOpts []alert.Option
	if dlgWidth := field.Tag.Get("dialogOptions"); dlgWidth != "" {
		switch dlgWidth {
		case "large":
			dlgOpts = append(dlgOpts, alert.Large())
		case "larger":
			dlgOpts = append(dlgOpts, alert.Larger())
		case "xlarge":
			dlgOpts = append(dlgOpts, alert.XLarge())
		case "xxlarge":
			dlgOpts = append(dlgOpts, alert.XXLarge())
		}
	}

	return dlgOpts
}
