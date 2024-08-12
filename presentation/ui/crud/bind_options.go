package crud

import (
	"errors"
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

type ElementViewFactory[E any] func(*E) core.View

func ButtonDelete[E any](wnd core.Window, deleteFn func(*E) error) ElementViewFactory[E] {
	return func(e *E) core.View {
		noSuchPermisionPresented := core.AutoState[bool](wnd)
		areYouSurePresented := core.AutoState[bool](wnd)
		return ui.VStack(
			alert.Dialog("Löschen nicht möglich", ui.Text("Sie sind nicht berechtigt diesen Eintrag zu löschen."), noSuchPermisionPresented, alert.Ok()),
			alert.Dialog("Bestätigung", ui.Text("Soll der Eintrag wirklich gelöscht werden?"), areYouSurePresented, alert.Cancel(nil), alert.Delete(func() {
				if err := deleteFn(e); err != nil {
					var denied PermissionDenied
					if errors.As(err, &denied) {
						noSuchPermisionPresented.Set(true)
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

// Views creates a field binding to E and renders with the binded E the given options.
// Keep in mind, to remove Render* functions, if it does not make sense or may cause
// malfunctions in the context, e.g. deleting an E without navigation.
func Views[E any](label string, options ...ElementViewFactory[E]) Field[E] {
	return Field[E]{
		Label: label,
		RenderTableCell: func(self Field[E], entity *E) ui.TTableCell {
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
		RenderCardElement: func(self Field[E], entity *E) ui.DecoredView {
			return ui.HStack(slices.Collect(func(yield func(cell core.View) bool) {
				for _, option := range options {
					yield(option(entity))
				}
			})...).
				Gap(ui.L4).
				Alignment(ui.Leading)
		},

		RenderFormElement: func(self Field[E], entity *E) ui.DecoredView {
			return self.RenderCardElement(self, entity)
		},
	}
}
