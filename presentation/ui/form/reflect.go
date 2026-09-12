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
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strings"

	"go.wdy.de/nago/application/xerror"
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

	// Errors is the validation result to display. Field bound messages are taken from every
	// [xerrors.ErrorWithFields] in the error tree, including those joined with errors.Join,
	// and are matched against the rendered fields. A key may either be the plain struct field
	// name ("Country") or the full path of a nested field ("Issuer.Seat.Country").
	//
	// Nothing is ever dropped: a message whose key matches no rendered field, and any part of
	// the error which is not field bound, is shown as an [alert.BannerError]. So a typo in a
	// key becomes visible instead of silently closing the form without feedback.
	//
	// See also [xerrors.WithFields] and [xerrors.FieldBuilder] to create a compatible error,
	// and [xerrors.FieldBuilder.Nested] for nested models.
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

// maxNestingDepth bounds the descent into nested struct fields, so that a self referencing
// model cannot turn the render into an infinite recursion.
const maxNestingDepth = 10

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

	// Collect every field bound message in the whole error tree. Using errors.AsType here
	// would stop at the first match and silently drop the field messages of all siblings of
	// an errors.Join.
	fieldErrors, _ := xerrors.Collect(t.opts.Errors)
	consumed := &consumedKeys{}

	var rootViews xslices.Builder[core.View]
	structType := reflect.TypeOf(value)
	rIt := t.opts.Renderers
	if rIt == nil {
		rIt = Renderers
	}

	if structType == nil || structType.Kind() != reflect.Struct {
		return ui.VStack(alert.Banner("implementation error", fmt.Sprintf("form.Auto requires a struct but got %v", structType))).Render(ctx)
	}

	pass := renderPass{ctx: ctx, renderers: rIt, fieldErrors: fieldErrors.Fields, consumed: consumed}

	for _, group := range LocalizeGroups(ctx.Window().Bundle(), GroupsOf(structType, t.opts.IgnoreFields...)) {
		fields := t.renderFields(pass, group.Fields, 0)

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

	// Whatever could not be attached to a visible field must still reach the user. Previously
	// the presence of any ErrorWithFields anywhere in the tree suppressed the banner outright,
	// so a typo in a field name, a joined infrastructure error, or a field whose renderer does
	// not display error texts made the whole error vanish without a trace.
	if residual := t.residualError(ctx.Window(), structType, consumed, fieldErrors.Fields); residual != nil {
		rootViews.Append(alert.BannerError(residual))
	}

	return ui.VStack(rootViews.Collect()...).Gap(ui.L16).FullWidth().Render(ctx)
}

// renderFields renders the given fields, descending into named nested struct fields which no
// renderer claims.
//
// The descent is decided here and not in [GroupsOf] on purpose: only here is the renderer set
// known. A type such as time.Time is claimed by a renderer and must never be descended into,
// and maintaining a list of such types elsewhere would be a permanent source of drift.
// renderPass holds everything that is constant for one render of the whole form, so that the
// recursive descent does not have to thread half a dozen unchanging parameters through every
// frame.
type renderPass struct {
	ctx         core.RenderContext
	renderers   iter.Seq[Renderer]
	fieldErrors map[string]string
	consumed    *consumedKeys
}

// descendable reports whether walkFields would look inside this field when no renderer claims
// it.
//
// Note that Go forbids a struct from containing itself by value, so the descent cannot cycle
// and no visited set is needed; the depth cap alone bounds it. Embedded structs are excluded
// because reflect.VisibleFields already promotes their leaves, and descending would render
// every one of them twice. Pointers are excluded because FieldByIndex cannot traverse a nil
// pointer.
func descendable(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Struct && !field.Anonymous
}

// walkFields visits every field that the auto form can address, descending into nested models.
//
// visit receives the field with its Index rewritten to be relative to the root model, together
// with the dotted path. Returning true means the field was handled and must not be descended
// into; returning false lets the walk look inside it if it is a nested model.
//
// This is the single definition of the addressing scheme. [TAuto.Render] and [FieldPaths] both
// go through it, so a form can never disagree with the paths it documents.
type fieldVisitor struct {
	// visit is called for every field. Returning true means handled, returning false lets the
	// walk descend into it if it is a nested model.
	visit func(field reflect.StructField, path string, depth int) bool
	// enter and leave bracket the fields of a nested model, so that a consumer can group them.
	enter func(field reflect.StructField, path string, depth int)
	leave func(field reflect.StructField, path string, depth int)
}

func walkFields(
	fields []reflect.StructField,
	ignoreFields []string,
	indexPrefix []int,
	pathPrefix string,
	depth int,
	v fieldVisitor,
) {
	for _, field := range fields {
		resolved := field
		if len(indexPrefix) > 0 {
			resolved.Index = append(append([]int{}, indexPrefix...), field.Index...)
		}

		path := field.Name
		if pathPrefix != "" {
			path = pathPrefix + xerrors.PathSeparator + field.Name
		}

		if v.visit(resolved, path, depth) {
			continue
		}

		if !descendable(field) {
			continue
		}

		if depth >= maxNestingDepth {
			slog.Warn("form: maximum nesting depth reached, field is not shown", "field", path, "depth", depth)
			continue
		}

		if v.enter != nil {
			v.enter(resolved, path, depth)
		}

		for _, group := range GroupsOf(field.Type, ignoreFields...) {
			walkFields(group.Fields, ignoreFields, resolved.Index, path, depth+1, v)
		}

		if v.leave != nil {
			v.leave(resolved, path, depth)
		}
	}
}

// FieldPaths returns every error key the auto form can address for the given model type, in
// render order, for example "Title" and "Issuer.Seat.Country".
//
// Use it in a test to assert that the keys a validation produces actually exist. A key that is
// not in this list never reaches a field and ends up in the banner instead.
func FieldPaths(modelType reflect.Type, ignoreFields ...string) []string {
	var out []string

	for _, group := range GroupsOf(modelType, ignoreFields...) {
		walkFields(group.Fields, ignoreFields, nil, "", 0, fieldVisitor{
			visit: func(field reflect.StructField, path string, depth int) bool {
				if descendable(field) && depth < maxNestingDepth {
					// let the walk descend, the container itself is not addressable
					return false
				}

				if field.Anonymous && field.Type.Kind() == reflect.Struct {
					return true
				}

				out = append(out, path)
				return true
			},
		})
	}

	return out
}

// renderFields renders the given fields, descending into named nested struct fields which no
// renderer claims.
//
// The descent is decided here and not in [GroupsOf] on purpose: only here is the renderer set
// known. A type such as time.Time is claimed by a renderer and must never be descended into,
// and maintaining a list of such types elsewhere would be a permanent source of drift.
func (t TAuto[T]) renderFields(p renderPass, fields []reflect.StructField, depth int) []core.View {
	// A stack of builders: the bottom one collects the root fields, and every descent pushes
	// one so that the nested fields can be wrapped into their own fieldset on the way out.
	stack := []*xslices.Builder[core.View]{{}}
	top := func() *xslices.Builder[core.View] { return stack[len(stack)-1] }

	walkFields(fields, t.opts.IgnoreFields, nil, "", depth, fieldVisitor{
		visit: func(field reflect.StructField, path string, depth int) bool {
			fctx := newFieldContext(p.ctx.Window(), t.opts, t.state, field, path, depth == 0, p.fieldErrors, p.consumed)

			for renderer := range p.renderers {
				if v := renderer(fctx); v != nil {
					top().Append(v)
					return true
				}
			}

			if descendable(field) {
				// no renderer is responsible, so let walkFields look inside instead of
				// dropping the field, which is what happened before
				return false
			}

			// reflect.VisibleFields reports the embedding struct itself in addition to its
			// promoted leaves. Those leaves are rendered individually, so there is nothing to
			// do and nothing to warn about here.
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				return true
			}

			slog.Warn("form: no renderer for field, it is not shown", "field", path, "type", field.Type.String())
			return true
		},

		enter: func(field reflect.StructField, path string, depth int) {
			stack = append(stack, &xslices.Builder[core.View]{})
		},

		leave: func(field reflect.StructField, path string, depth int) {
			views := top().Collect()
			stack = stack[:len(stack)-1]

			if len(views) == 0 {
				return
			}

			label := field.Name
			if name, ok := field.Tag.Lookup("label"); ok {
				label = name
			}

			top().Append(
				Fieldset(ui.VStack(views...).Gap(ui.L16).FullWidth()).
					Title(p.ctx.Window().Bundle().Resolve(label)).
					Frame(ui.Frame{}.FullWidth()),
			)
		},
	})

	return top().Collect()
}

// residualError returns what still has to be shown as a banner after the field bound messages
// were rendered on the fields themselves.
//
// The invariant is that nothing the caller passed in disappears. Two things can remain:
// the non field part of the error tree, and messages addressed to a key that no rendered
// field claimed.
func (t TAuto[T]) residualError(wnd core.Window, modelType reflect.Type, consumed *consumedKeys, fieldErrors map[string]string) error {
	if t.opts.Errors == nil {
		return nil
	}

	// Everything that is not field bound, for example an infrastructure error joined to a
	// validation error. Collect alone would make the whole thing look handled.
	residual := xerrors.Residual(t.opts.Errors)

	var unclaimed []string
	for _, key := range slices.Sorted(maps.Keys(fieldErrors)) {
		if !consumed.has(key) {
			unclaimed = append(unclaimed, key+": "+fieldErrors[key])
		}
	}

	if len(unclaimed) == 0 {
		return residual
	}

	// A message was addressed to a field which is not rendered: a typo in the key, a field
	// hidden by IgnoreFields, or a renderer which does not display error texts. The key is a
	// Go identifier and therefore a developer diagnostic, so it goes to the log while the
	// user gets the generic hint. What must not happen is what happened before: nothing at
	// all, leaving the form neither closing nor explaining itself.
	slog.Error("form: validation messages did not reach any rendered field",
		"unclaimed", strings.Join(unclaimed, "; "),
		"model", fmt.Sprintf("%v", modelType),
		"addressable", strings.Join(FieldPaths(modelType, t.opts.IgnoreFields...), ", "),
	)

	// wnd may be nil when the form is rendered outside a window, so route through the safe
	// bundler rather than dereferencing it.
	b := xerror.BundlerOrDefault(wnd)

	hint := std.NewLocalizedError(
		xerror.StrValidationFailed.Get(b),
		xerror.StrValidationFailedMsg.Get(b),
	)

	if residual == nil {
		return hint
	}

	return errors.Join(residual, hint)
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
