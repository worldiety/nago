// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func (c TFormEditor) viewHelpFieldsStmt(structType *flow.StructType, state *core.State[string]) core.View {
	var views []core.View
	views = append(views, ui.Text("Field helper templates: click a value to append it as statement"))
	for field := range structType.Fields.All() {
		headingName := field.JSONName()
		if headingName != string(field.Name()) {
			headingName += " (" + string(field.Name()) + ")"
		}

		views = append(views, ui.VStack(
			ui.Heading(5, headingName),
			c.possibleFieldValuesPut(field, state),
		).Alignment(ui.Leading))
	}

	return ui.VStack(views...).Alignment(ui.Leading).FullWidth().Gap(ui.L8)
}

func (c TFormEditor) possibleFieldValuesPut(field flow.Field, state *core.State[string]) core.View {
	var views []core.View
	switch field := field.(type) {
	case *flow.StringField:
		views = append(views, pillTextPutStmt(state, field, mfVEmpty), pillTextPutStmt(state, field, mfVDelete), pillTextPutStmt(state, field, `any text`))
	case *flow.BoolField:
		views = append(views, pillTextPutStmt(state, field, mfVTrue), pillTextPutStmt(state, field, mfVFalse), pillTextPutStmt(state, field, mfVDelete))
	case *flow.TypeField:
		baseType, ok := c.ws.Packages.TypeByID(field.Type)
		if !ok {
			views = append(views, pillTextPutStmt(state, field, "unknown type"))
		} else {
			if str, ok := baseType.(*flow.StringType); ok {
				views = append(views, pillTextPutStmt(state, field, mfVEmpty), pillTextPutStmt(state, field, mfVDelete))
				if str.Enumeration.Len() == 0 {
					views = append(views, pillTextPutStmt(state, field, `any text`))
				}

				for literal := range str.Enumeration.All() {
					views = append(views, pillTextPutStmt(state, field, literal.Value))
				}
			}
		}
	}
	return ui.HStack(views...).Gap(ui.L8).Wrap(true)
}

func pillTextPutStmt(state *core.State[string], field flow.Field, s string) core.View {
	return ui.HStack(ui.Text(s)).
		BackgroundColor(ui.ColorCardFooter).
		Action(func() {
			fname := field.JSONName()
			var expr string
			switch s {
			case mfVEmpty:
				expr = fmt.Sprintf(`put("%s","")`, fname)
			case mfVTrue:
				expr = fmt.Sprintf(`put("%s",true)`, fname)
			case mfVFalse:
				expr = fmt.Sprintf(`put("%s",false)`, fname)
			case mfVDelete:
				expr = fmt.Sprintf(`delete("%s")`, fname)
			default:
				expr = fmt.Sprintf(`put("%s","%s")`, fname, s)
			}

			state.Set(strings.TrimSpace(state.Get()))
			if state.Get() != "" {
				expr = "\n" + expr
			}

			state.Set(state.Get() + expr)
		}).
		Padding(ui.Padding{}.All(ui.L4).Horizontal(ui.L8)).
		Border(ui.Border{}.Radius(ui.L8))
}
