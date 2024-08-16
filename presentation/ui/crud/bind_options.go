package crud

import (
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/data"
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
	wnd := bnd.wnd
	// we have a unique state problem here, at least we can have multiple crud views with distinct types
	var typ = fmt.Sprintf("%T", initial)
	entityState := core.StateOf[E](wnd, fmt.Sprintf("crud.create.entity.%s", typ)).From(func() E {
		return initial
	})
	formPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.create.form.%s", typ))
	noSuchPermissionPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.create.npp.%s", typ))
	stateErrMsg := core.StateOf[string](wnd, fmt.Sprintf("crud.create.errmsg.%s", typ))
	return ui.VStack(
		alert.Dialog("Erstellen nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu erstellen."), noSuchPermissionPresented, alert.Ok()),
		alert.Dialog("Neu erstellen", ui.Composable(func() core.View {
			subBnd := bnd.Inherit(string(initial.Identity()))

			return ui.VStack(
				ui.If(stateErrMsg.Get() != "", ui.Text(stateErrMsg.Get()).Color(ui.SE0)),
				Form(subBnd, entityState),
			).Frame(ui.Frame{}.FullWidth())
		}), formPresented, alert.Cancel(func() {
			entityState.Set(initial)
		}), alert.Save(func() bool {
			errMsg, err := createFn(entityState.Get())
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

			entityState.Set(initial)
			return true
		})),

		ui.PrimaryButton(func() {
			formPresented.Set(true)
		},
		).PreIcon(heroSolid.Plus).AccessibilityLabel("Hinzufügen"))
}

func ButtonEdit[E data.Aggregate[ID], ID data.IDType](bnd *Binding[E], updateFn func(E) (errorText string, infrastructureError error)) ElementViewFactory[E] {
	wnd := bnd.wnd
	return func(e *core.State[E]) core.View {
		entityState := core.StateOf[E](wnd, fmt.Sprintf("crud.edit.entity.%v", e.Get().Identity())).From(func() E {
			return e.Get()
		})

		formPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.edit.form.%v", e.Get().Identity()))
		noSuchPermissionPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.edit.npp.%v", e.Get().Identity()))
		stateErrMsg := core.StateOf[string](wnd, fmt.Sprintf("crud.edit.errmsg.%v", e.Get().Identity()))
		return ui.VStack(
			alert.Dialog("Bearbeiten nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu bearbeiten."), noSuchPermissionPresented, alert.Ok()),
			alert.Dialog("Bearbeiten", ui.Composable(func() core.View {
				subBnd := bnd.Inherit(string(e.Get().Identity()))

				return ui.VStack(
					ui.If(stateErrMsg.Get() != "", ui.Text(stateErrMsg.Get()).Color(ui.SE0)),
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
			})),
			ui.PrimaryButton(func() {
				formPresented.Set(true)
			}).PreIcon(heroSolid.Pencil).AccessibilityLabel("Bearbeiten"),
		)
	}
}

func ButtonDelete[E data.Aggregate[ID], ID data.IDType](wnd core.Window, deleteFn func(E) error) ElementViewFactory[E] {
	return func(e *core.State[E]) core.View {
		noSuchPermissionPresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.delete.npp.%v", e.Get().Identity()))
		areYouSurePresented := core.StateOf[bool](wnd, fmt.Sprintf("crud.delete.sure.%v", e.Get().Identity()))
		return ui.VStack(
			alert.Dialog("Löschen nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu löschen."), noSuchPermissionPresented, alert.Ok()),
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
			ui.PrimaryButton(func() {
				areYouSurePresented.Set(true)

			}).PreIcon(heroSolid.Trash).AccessibilityLabel("Löschen"),
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
