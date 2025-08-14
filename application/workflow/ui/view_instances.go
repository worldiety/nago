// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiworkflow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/workflow"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/pager"
)

func specInstances(wnd core.Window, uc workflow.UseCases, id workflow.ID) core.View {
	model, err := pager.NewModel(
		wnd,
		func(id workflow.Instance) (option.Opt[workflow.Status], error) {
			return uc.GetStatus(wnd.Subject(), id)
		},
		uc.FindInstances(wnd.Subject(), id),
		pager.ModelOptions{},
	)

	if err != nil {
		return alert.BannerError(err)
	}

	analyzed := core.AutoState[workflow.Analyzed](wnd).Init(func() workflow.Analyzed {
		// keep some in-memory cache, perhaps quite expensive all those iterations
		res, err := uc.Analyze(wnd.Subject(), id)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return res
		}

		return res
	})

	showExternalEvents := core.AutoState[bool](wnd)
	selectedEvent := core.AutoState[reflect.Type](wnd)
	selectedEventJSON := core.AutoState[string](wnd)
	selectedEventDialogShown := core.AutoState[bool](wnd)

	return ui.VStack(
		func() core.View {
			if !selectedEventDialogShown.Get() {
				return nil
			}

			return alert.Dialog(
				"Event-Daten definieren",
				ui.VStack(
					ui.H2(string(workflow.NewTypename(selectedEvent.Get()))),
					ui.CodeEditor(selectedEventJSON.Get()).
						InputValue(selectedEventJSON).
						Language("json").
						FullWidth(),
				).Alignment(ui.Leading).
					FullWidth(),
				selectedEventDialogShown,
				alert.Closeable(),
				alert.Larger(),
				alert.Cancel(nil),
				alert.Custom(func(close func(closeDlg bool)) core.View {
					return ui.PrimaryButton(
						func() {
							// note that obj is a * to type which we need to keep, otherwise unmarshal looses the itself
							obj := reflect.New(selectedEvent.Get()).Interface()
							if err := json.Unmarshal([]byte(selectedEventJSON.Get()), &obj); err != nil {
								alert.ShowBannerMessage(wnd, alert.Message{
									Title:   "Invalid JSON",
									Message: "Das JSON ist ungültig und kann nicht geparsed werden: " + err.Error(),
								})
								return
							}

							slog.Info("issue manual event", "value", fmt.Sprintf("%+v", obj))

							if err := uc.ProcessEvent(wnd.Subject(), reflect.ValueOf(obj).Elem().Interface()); err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}

							selectedEventDialogShown.Set(false)
						},
					).Title("Event senden")
				}),
			)
		}(),

		alert.Dialog(
			"Event-Typ wählen",
			list.List(
				ui.ForEach(analyzed.Get().StartEvents, func(t workflow.Typename) core.View {
					name := analyzed.Get().EventTypesByName[t].String()
					return list.Entry().Leading(ui.Text(name)).Trailing(ui.ImageIcon(icons.ArrowRight))
				})...,
			).OnEntryClicked(func(idx int) {
				tname := analyzed.Get().StartEvents[idx]
				typ := analyzed.Get().EventTypesByName[tname]

				buf, err := json.MarshalIndent(reflect.New(typ).Interface(), "", "  ")
				if err != nil {
					slog.Error("cannot marshal template for event", "err", err.Error(), "type", tname)
				}

				if len(buf) == 0 {
					buf = []byte("{\n}")
				}

				selectedEventJSON.Set(string(buf))
				selectedEvent.Set(typ)
				selectedEventDialogShown.Set(true)
				showExternalEvents.Set(false)
			}),
			showExternalEvents,
			alert.Closeable(),
			alert.Close(nil),
		),
		ui.HStack(
			ui.PrimaryButton(func() {
				showExternalEvents.Set(true)
			}).Title("Globales Ereignis auslösen"),
		).FullWidth().
			Alignment(ui.Trailing),

		ui.Space(ui.L32),

		ui.Table(
			ui.TableColumn(ui.Checkbox(model.SelectSubset.Get()).InputChecked(model.SelectSubset)).Width(ui.L64),
			ui.TableColumn(ui.Text("Gestartet um")),
			ui.TableColumn(ui.Text("Gestoppt um")),
			ui.TableColumn(ui.Text("Status")),
			ui.TableColumn(ui.Text("")).Width(ui.L64),
		).Rows(
			ui.ForEach(model.Page.Items, func(u workflow.Status) ui.TTableRow {
				myState := model.Selections[u.ID]

				return ui.TableRow(
					ui.TableCell(ui.Checkbox(myState.Get()).InputChecked(myState)),
					ui.TableCell(ui.Text(xtime.FormatDateTime(wnd.Locale(), u.StartedAt))),
					ui.TableCell(ui.Text(xtime.FormatDateTime(wnd.Locale(), u.StoppedAt))),
					ui.TableCell(ui.Text(u.State.String())),
					ui.TableCell(ui.ImageIcon(icons.ChevronRight)),
				).Action(func() {
					wnd.Navigation().ForwardTo("admin/workflow/instance/events", core.Values{"id": string(id), "instance": string(u.ID)})
				}).HoveredBackgroundColor(ui.ColorCardFooter)
			})...,
		).Rows(
			ui.TableRow(
				ui.TableCell(
					ui.HStack(
						ui.Text(model.PageString()),
						ui.Spacer(),
						pager.Pager(model.PageIdx).Count(model.Page.PageCount).Visible(model.HasPages()),
					).FullWidth(),
				).ColSpan(6),
			).BackgroundColor(ui.ColorCardFooter),
		).
			Frame(ui.Frame{}.FullWidth()),
	).FullWidth()
}
