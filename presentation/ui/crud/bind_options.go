package crud

import (
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/tracking"
)

type PermissionDenied interface {
	error
	PermissionDenied() bool
}

type ElementViewFactory[E any] func(*core.State[E]) core.View

func ButtonEdit[E data.Aggregate[ID], ID data.IDType](wnd core.Window, bnd *Binding[E], updateFn func(E) (errorText string, infrastructureError error)) ElementViewFactory[E] {
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

// OptionsField is like [Views] but omits the views from the form to avoid render recursion.
func OptionsField[E any](label string, options ...ElementViewFactory[E]) Field[E] {
	return Views[E](label, options...).WithoutForm()
}

// Views creates a field binding to E and renders with the bound E the given options.
// Keep in mind, to remove Render* functions, if it does not make sense or may cause
// malfunctions in the context, e.g. deleting an E without navigation.
func Views[E any](label string, options ...ElementViewFactory[E]) Field[E] {
	return Field[E]{
		Label: label,
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			return ui.TableCell(ui.HStack(slices.Collect(func(yield func(cell core.View) bool) {
				for _, option := range options {
					yield(option(entity))
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
