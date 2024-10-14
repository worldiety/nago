package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
)

// Section takes a bunch of fields and puts them into a common card on the form.
// However, the table and card rendering is nothing special and just returned.
// To make that more clear, consider the fields A, B and C which shall be put into a section.
// The returned Fields are F, A', B' and C'. They are created as follows:
//
//	F.RenderFormElement = SectionView(A.RenderFormElement, B.RenderFormElement, C.RenderFormElement)
//
//	A'.RenderTableCell = A.RenderTableCell
//	A'.RenderCardElement = A.RenderCardElement
//	A'.RenderFormElement = nil
//
//	B'.RenderTableCell = B.RenderTableCell
//	B'.RenderCardElement = B.RenderCardElement
//	B'.RenderFormElement = nil
//
//	C'.RenderTableCell = C.RenderTableCell
//	C'.RenderCardElement = C.RenderCardElement
//	C'.RenderFormElement = nil
func Section[E any](label string, fields ...Field[E]) []Field[E] {
	return fakeFormFields(label, section, fields...)
}

func fakeFormFields[E any](label string, render func(views ...core.View) ui.DecoredView, fields ...Field[E]) []Field[E] {
	modifiedFields := make([]Field[E], 0, len(fields))
	formRenderers := make([]func(self Field[E], entity *core.State[E]) ui.DecoredView, 0, len(fields))
	for _, field := range fields {
		formRenderers = append(formRenderers, field.RenderFormElement)
		field.RenderFormElement = nil
		if field.metaRefID == "" {
			field.metaRefID = data.RandIdent[string]()
		}

		modifiedFields = append(modifiedFields, field)
	}

	res := make([]Field[E], 0, len(fields)+1)
	res = append(res, Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			views := make([]core.View, 0, len(formRenderers)+1)
			if label != "" {
				views = append(views, ui.VStack(
					ui.Text(label).Font(ui.Title),
					ui.HLine(),
				))
			}

			for idx, modField := range modifiedFields {
				var field Field[E]
				for _, f := range self.parent.fields {
					if f.metaRefID == modField.metaRefID {
						field = f
						break
					}
				}

				if field.metaRefID == "" {
					slog.Error("crud.Section cannot find field by meta ref id")
					continue
				}

				if renderer := formRenderers[idx]; renderer != nil {
					views = append(views, renderer(field, entity))
				}
			}

			return render(views...)
		},
	})

	res = append(res, modifiedFields...)
	return res
}

func section(views ...core.View) ui.DecoredView {
	return ui.VStack(
		views...,
	).Alignment(ui.TopLeading).Gap(ui.L12).BackgroundColor(ui.ColorCardBody).Border(ui.Border{
		TopLeftRadius:     ui.L8,
		TopRightRadius:    ui.L8,
		BottomLeftRadius:  ui.L8,
		BottomRightRadius: ui.L8,
	}).Padding(ui.Padding{}.All("1rem")).Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{
			TopLeftRadius:     ui.L16,
			TopRightRadius:    ui.L16,
			BottomLeftRadius:  ui.L16,
			BottomRightRadius: ui.L16,
		})

}
