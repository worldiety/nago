// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/tracking"
	"slices"
)

type PermissionDenied interface {
	error
	PermissionDenied() bool
}

type ElementViewFactory[E any] func(*core.State[E]) core.View

// Optional makes an ElementViewFactory optional based on the given predicate and returns nil if predicate returns false.
func Optional[T any](fac ElementViewFactory[T], predicate func(T) bool) ElementViewFactory[T] {
	return func(c *core.State[T]) core.View {
		if predicate(c.Get()) {
			return fac(c)
		}

		return nil
	}
}

func ButtonCreate[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], initial E, createFn func(E) (errorText string, infrastructureError error)) core.View {
	dlg, formPresented := CreateDialog(bnd, initial, createFn, nil, nil)
	return ui.VStack(
		dlg,
		ui.PrimaryButton(func() {
			formPresented.Set(true)
		},
		).PreIcon(heroSolid.Plus).Title(xstrings.Join2(" ", xstrings.If(bnd.entityAliasName == "", "Eintrag", bnd.entityAliasName), "hinzufügen")))
}

func ButtonCreateForwardTo[E any](bnd *Binding[E], path core.NavigationPath, values core.Values) core.View {
	return ui.VStack(
		ui.PrimaryButton(func() {
			bnd.wnd.Navigation().ForwardTo(path, values)
		},
		).PreIcon(heroSolid.Plus).Title(xstrings.Join2(" ", xstrings.If(bnd.entityAliasName == "", "Eintrag", bnd.entityAliasName), "hinzufügen")))
}

// CreateDialog returns the dialog for creating. See also [DialogEdit].
func CreateDialog[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], initial E, createFn func(E) (errorText string, infrastructureError error), onCancelled, onSaved func()) (dlg core.View, presented *core.State[bool]) {
	wnd := bnd.wnd
	// we have a unique state problem here, at least we can have multiple crud views with distinct types
	var typ = fmt.Sprintf("%T", initial)
	entityState := core.StateOf[E](wnd, fmt.Sprintf("crud.create.entity.%s", typ)).Init(func() E {
		return initial
	})
	formPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.create.form.%s", typ))
	noSuchPermissionPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.create.npp.%s", typ))
	stateErrMsg := core.StateOf[string](wnd, fmt.Sprintf("crud.create.errmsg.%s", typ))
	errState := core.StateOf[error](wnd, fmt.Sprintf("crud.create.error.%s", typ))
	return ui.VStack(

		alert.Dialog(xstrings.Join2(" ", xstrings.If(bnd.entityAliasName == "", "Eintrag", bnd.entityAliasName), "erstellen"), ui.Composable(func() core.View {
			subBnd := bnd.Inherit(data.Idtos(initial.Identity()))

			return ui.VStack(
				alert.BannerError(errState.Get()),
				ui.If(stateErrMsg.Get() != "", ui.Text(stateErrMsg.Get()).Color(ui.SE0)),
				Form(subBnd, entityState),
			).Frame(ui.Frame{}.FullWidth())
		}), formPresented, alert.Cancel(func() {
			bnd.ResetValidation()
			stateErrMsg.Set("")
			entityState.Set(initial)
			if onCancelled != nil {
				onCancelled()
			}
		}), alert.Save(func() bool {
			errMsg, err := createFn(entityState.Get())
			if err != nil {
				errState.Set(err)

				return false
			}

			errState.Set(nil)

			stateErrMsg.Set(errMsg)
			if errMsg != "" {
				return false
			}

			entityState.Set(initial)
			stateErrMsg.Set("")
			bnd.ResetValidation()
			if onSaved != nil {
				onSaved()
			}
			return true
		}),
			alert.Width(ui.L560),
		),
		alert.Dialog("Erstellen nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu erstellen."), noSuchPermissionPresented, alert.Ok()),
	), formPresented

}

func RenderElementViewFactory[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], entity E, fac ElementViewFactory[E]) core.View {
	bnd = bnd.Inherit(data.Idtos(entity.Identity()))
	entityState := core.StateOf[E](bnd.wnd, fmt.Sprintf("crud.renderfactory.entity.%v", entity.Identity())).Init(func() E {
		return entity
	})
	entityState.Set(entity)

	return fac(entityState)
}

// DialogEdit returns the dialog for editing. See also [CreateDialog].
func DialogEdit[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], presented *core.State[bool], entity E, updateFn func(E) (errorText string, infrastructureError error)) core.View {
	return RenderElementViewFactory(bnd, entity, buttonEdit(bnd, false, presented, updateFn))
}

// ButtonEdit to be used for conventional delete function. See also [ViewButtonEdit] for other use cases.
func ButtonEdit[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], updateFn func(E) (errorText string, infrastructureError error)) ElementViewFactory[E] {
	wnd := bnd.wnd
	return func(e *core.State[E]) core.View {
		formPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.edit.form.%v", e.Get().Identity()))
		return buttonEdit(bnd, true, formPresented, updateFn)(e)
	}
}

func ButtonEditForwardTo[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], navigate func(wnd core.Window, entity E)) ElementViewFactory[E] {
	return func(e *core.State[E]) core.View {
		return ui.TertiaryButton(func() {
			navigate(bnd.wnd, e.Get())
		}).PreIcon(heroSolid.Pencil).AccessibilityLabel("Bearbeiten")
	}
}

func buttonEdit[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], renderBtn bool, formPresented *core.State[bool], updateFn func(E) (errorText string, infrastructureError error)) ElementViewFactory[E] {
	wnd := bnd.wnd
	return func(e *core.State[E]) core.View {
		entityState := core.StateOf[E](wnd, fmt.Sprintf("crud.edit.entity.%v", e.Get().Identity())).Init(func() E {
			return e.Get()
		})

		noSuchPermissionPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.edit.npp.%v", e.Get().Identity()))
		stateErrMsg := core.StateOf[string](wnd, fmt.Sprintf("crud.edit.errmsg.%v", e.Get().Identity()))
		return ui.VStack(

			alert.Dialog(xstrings.Join2(" ", xstrings.If(bnd.entityAliasName == "", "Eintrag", bnd.entityAliasName), "bearbeiten"), ui.Composable(func() core.View {
				subBnd := bnd.Inherit(data.Idtos(e.Get().Identity()))

				return ui.VStack(
					ui.If(stateErrMsg.Get() != "", ui.Text(stateErrMsg.Get()).Color(ui.SE0)),
					ui.If(bnd.deleteFunc != nil,
						ui.HStack(RenderElementViewFactory[E, ID](bnd, e.Get(), ButtonDeleteWithCaption[E, ID](bnd.wnd, xstrings.Join2(" ", bnd.entityAliasName, "löschen"), bnd.deleteFunc))).
							FullWidth().
							Alignment(ui.Trailing),
					),
					Form(subBnd, entityState),
				).Frame(ui.Frame{}.FullWidth())
			}), formPresented, alert.Cancel(func() {
				entityState.Set(e.Get())
			}), alert.Save(func() bool {
				errMsg, err := updateFn(entityState.Get())
				if err != nil {
					var denied PermissionDenied
					if errors.As(err, &denied) {
						noSuchPermissionPresented.Set(true)
					} else {
						tracking.RequestSupport(wnd, err)
					}

					return false
				}

				stateErrMsg.Set(errMsg)
				if errMsg != "" {
					return false
				}

				return true
			}),
				alert.Width(ui.L560),
			),
			ui.If(renderBtn, ui.TertiaryButton(func() {
				formPresented.Set(true)
			}).PreIcon(heroSolid.Pencil).AccessibilityLabel("Bearbeiten")),

			alert.Dialog("Bearbeiten nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu bearbeiten."), noSuchPermissionPresented, alert.Ok()),
		)
	}
}
func ButtonDelete[E data.Aggregate[ID], ID data.IDType](wnd core.Window, deleteFn func(E) error) ElementViewFactory[E] {
	return ButtonDeleteWithCaption(wnd, "", deleteFn)
}

func ButtonDeleteWithCaption[E data.Aggregate[ID], ID data.IDType](wnd core.Window, caption string, deleteFn func(E) error) ElementViewFactory[E] {
	return func(e *core.State[E]) core.View {
		noSuchPermissionPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.delete.npp.%v", e.Get().Identity()))
		areYouSurePresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.delete.sure.%v", e.Get().Identity()))
		return ui.VStack(
			alert.Dialog("Bestätigung", ui.Text("Soll der Eintrag wirklich gelöscht werden?"), areYouSurePresented, alert.Cancel(nil), alert.Delete(func() {
				if err := deleteFn(e.Get()); err != nil {
					var denied PermissionDenied
					if errors.As(err, &denied) {
						noSuchPermissionPresented.Set(true)
					} else {
						tracking.RequestSupport(wnd, err)
					}
				}
			})),
			tracking.SupportRequestDialog(wnd),
			ui.TertiaryButton(func() {
				areYouSurePresented.Set(true)

			}).PreIcon(heroSolid.Trash).AccessibilityLabel("Löschen").Title(caption),
			alert.Dialog("Löschen nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu löschen."), noSuchPermissionPresented, alert.Ok()),
		)

	}
}

// AggregateActions creates a field binding to E and renders with the bound E the given options.
// Keep in mind, to remove Render* functions, if it does not make sense or may cause
// malfunctions in the context, e.g. deleting an E without navigation.
// AggregateActions omits the views from the form to avoid render recursion.
func AggregateActions[E any](label string, options ...ElementViewFactory[E]) Field[E] {
	return rowInField[E](label, options...).WithoutForm()
}

func rowInField[E any](label string, options ...ElementViewFactory[E]) Field[E] {
	return Field[E]{
		Label: label,
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			return ui.TableCell(ui.HStack(slices.Collect(func(yield func(cell core.View) bool) {
				for _, option := range options {
					view := option(entity)
					if view != nil {
						yield(view)
					}
				}
			})...).
				// hstack
				Gap(ui.L4).
				Alignment(ui.Trailing)).
				//table cell
				Alignment(ui.Trailing)
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			return ui.HStack(slices.Collect(func(yield func(cell core.View) bool) {
				for _, option := range options {
					yield(option(entity))
				}
			})...).
				Gap(ui.L4).
				Alignment(ui.Leading)
		},

		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			return nil //self.RenderCardElement(self, entity)
		},
	}
}
