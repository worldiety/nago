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

// magic field values
const (
	mfVEmpty    = "[empty]"
	mfVNotEmpty = "[not empty]"
	mfVTrue     = "true"
	mfVFalse    = "false"
	mfVNull     = "[null]"
	mfVHas      = "[has]"
	mfVDelete   = "[delete]"
)

func (c TFormEditor) viewHelpFieldsExpr(structType *flow.StructType, state *core.State[string]) core.View {
	var views []core.View
	views = append(views, ui.Text("Field helper templates: click a value to append it as expression"))
	for field := range structType.Fields.All() {
		headingName := field.JSONName()
		if headingName != string(field.Name()) {
			headingName += " (" + string(field.Name()) + ")"
		}

		views = append(views, ui.VStack(
			ui.Heading(5, headingName),
			c.possibleFieldValuesExpr(field, state),
		).Alignment(ui.Leading))
	}

	return ui.VStack(views...).Alignment(ui.Leading).FullWidth().Gap(ui.L8)
}

func (c TFormEditor) possibleFieldValuesExpr(field flow.Field, state *core.State[string]) core.View {
	var views []core.View
	switch field := field.(type) {
	case *flow.StringField:
		views = append(views, pillTextExpr(state, field, mfVEmpty), pillTextExpr(state, field, mfVNotEmpty), pillTextExpr(state, field, mfVHas), pillTextExpr(state, field, `any text`))
	case *flow.BoolField:
		views = append(views, pillTextExpr(state, field, mfVTrue), pillTextExpr(state, field, mfVFalse), pillTextExpr(state, field, mfVHas))
	case *flow.TypeField:
		baseType, ok := c.ws.Packages.TypeByID(field.Type)
		if !ok {
			views = append(views, pillTextExpr(state, field, "unknown type"))
		} else {
			if str, ok := baseType.(*flow.StringType); ok {
				views = append(views, pillTextExpr(state, field, mfVEmpty), pillTextExpr(state, field, mfVNotEmpty), pillTextExpr(state, field, mfVHas))
				if str.Enumeration.Len() == 0 {
					views = append(views, pillTextExpr(state, field, `any text`))
				}

				for literal := range str.Enumeration.All() {
					views = append(views, pillTextExpr(state, field, literal.Value))
				}
			}
		}
	}
	return ui.HStack(views...).Gap(ui.L8).Wrap(true)
}

func pillTextExpr(state *core.State[string], field flow.Field, s string) core.View {
	return ui.HStack(ui.Text(s)).
		BackgroundColor(ui.ColorCardFooter).
		Action(func() {
			fname := field.JSONName()
			var expr string
			switch s {
			case mfVEmpty:
				expr = fmt.Sprintf(`field("%s").String() == ""`, fname)
			case mfVNotEmpty:
				expr = fmt.Sprintf(`field("%s").String() != ""`, fname)
			case mfVTrue:
				expr = fmt.Sprintf(`field("%s").Bool() == true`, fname)
			case mfVFalse:
				expr = fmt.Sprintf(`field("%s").Bool() == false`, fname)
			case mfVHas:
				expr = fmt.Sprintf(`has("%s")`, fname)
			default:
				expr = fmt.Sprintf(`field("%s").String() == "%s"`, fname, s)
			}

			state.Set(strings.TrimSpace(state.Get()))
			if state.Get() != "" {
				expr = " && " + expr
			}

			state.Set(state.Get() + expr)
		}).
		Padding(ui.Padding{}.All(ui.L4).Horizontal(ui.L8)).
		Border(ui.Border{}.Radius(ui.L8))
}
