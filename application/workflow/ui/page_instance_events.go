// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiworkflow

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/workflow"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
)

func PageInstanceEvents(wnd core.Window, uc workflow.UseCases) core.View {
	wid := workflow.ID(wnd.Values()["id"])

	optWf, err := uc.FindDeclaredWorkflow(wnd.Subject(), wid)
	if err != nil {
		return alert.BannerError(err)
	}

	if optWf.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	wf := optWf.Unwrap()

	instance := workflow.Instance(wnd.Values()["instance"])

	dialogJsonPresented := core.AutoState[bool](wnd)
	jsonData := core.AutoState[string](wnd)

	return ui.VStack(
		breadcrumb.Breadcrumbs(
			ui.TertiaryButton(func() {
				wnd.Navigation().ForwardTo("admin/workflow", core.Values{"id": string(wid), "pager-index-idx": "1"})
			}).Title(wf.Name),
			ui.TertiaryButton(nil).Title(string(instance)),
		),
		msgRawDialog(wnd, dialogJsonPresented, jsonData),
		dataview.FromData(wnd, dataview.Data[workflow.Event, workflow.EventKey]{
			FindAll: uc.FindInstanceEvents(wnd.Subject(), instance),
			FindByID: func(id workflow.EventKey) (option.Opt[workflow.Event], error) {
				return uc.FindInstanceEvent(wnd.Subject(), id)
			},
			Fields: []dataview.Field[workflow.Event]{
				{
					Name: "Nummer",
					Map: func(obj workflow.Event) core.View {
						return ui.Text(strconv.Itoa(int(obj.SeqNo)))
					},
				},

				{
					Name: "Uhrzeit",
					Map: func(obj workflow.Event) core.View {
						return ui.Text(xtime.FormatDateTime(wnd.Locale(), obj.SavedAt.Time(time.Local)))
					},
				},
				{
					Name: "Typ",
					Map: func(obj workflow.Event) core.View {
						return ui.Text(string(workflow.NewTypename(reflect.TypeOf(obj.Payload))))
					},
				},
				{
					Name: "",
					Map: func(obj workflow.Event) core.View {
						return ui.HStack(
							ui.SecondaryButton(func() {
								buf, err := json.MarshalIndent(obj, "", "  ")
								if err != nil {
									jsonData.Set(err.Error())
								} else {
									jsonData.Set(string(buf))
								}

								dialogJsonPresented.Set(true)
							}).Title("Event anzeigen"),
						).FullWidth().Alignment(ui.Trailing)

					},
				},
			},
		}),
	).Alignment(ui.Leading).
		Gap(ui.L16).
		FullWidth()
}

func msgRawDialog(wnd core.Window, presented *core.State[bool], data *core.State[string]) core.View {
	return alert.Dialog(
		"Event",
		ui.CodeEditor(data.Get()).
			Language("json").
			FullWidth(),
		presented,
		alert.Closeable(),
		alert.Larger(),
		alert.Close(nil),
	)
}
