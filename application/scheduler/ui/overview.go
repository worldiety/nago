// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uischeduler

import (
	"encoding/json"
	"go.wdy.de/nago/application/scheduler"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/form"
	"time"
)

func PageOverview(wnd core.Window, scheduleUseCases scheduler.UseCases) core.View {
	sid := scheduler.ID(wnd.Values()["id"])
	status, err := scheduleUseCases.Status(wnd.Subject(), sid)
	if err != nil {
		return alert.BannerError(err)
	}

	editSettingsPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1(status.Options.Name),
		ui.Text(status.Options.Description),
		ui.RedrawAtFixedRate[core.View](wnd, time.Second, nil),
		editSettingsDialog(wnd, status, editSettingsPresented, scheduleUseCases),
		ui.FixedSpacer(ui.L16, ui.L16),
		ui.HStack(
			ui.SecondaryButton(func() {
				if err := scheduleUseCases.Stop(wnd.Subject(), sid); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

			}).Title("Beenden"),
			ui.SecondaryButton(func() {
				if err := scheduleUseCases.Start(wnd.Subject(), sid); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}
			}).Title("Starten"),
			ui.SecondaryButton(func() {
				go func() {
					if err := scheduleUseCases.ExecuteNow(wnd.Subject(), sid); err != nil {
						alert.ShowBannerError(wnd, err)
					}
				}()
			}).Title("jetzt ausführen"),
		).FullWidth().Alignment(ui.Trailing).Gap(ui.L8),
		ui.FixedSpacer(ui.L24, ui.L24),
		cardlayout.Card("Übersicht").
			Body(ui.VStack(
				ui.HStack(
					ui.ImageIcon(kindIco(status.Options.Kind)),
					ui.Spacer(),
					ui.Text(kindStr(status.Options.Kind)),
				).FullWidth(),
				ui.HLine(),

				ui.HStack(ui.Text("Status"), ui.Spacer(), ui.Text(stateStr(status.State))).FullWidth(),
				ui.HLine(),

				ui.HStack(ui.Text("zuletzt gestartet"), ui.Spacer(), ui.Text(formatDate(status.LastStartedAt))).FullWidth(),
				ui.HLine(),

				ui.HStack(ui.Text("zuletzt durchlaufen"), ui.Spacer(), ui.Text(formatDate(status.LastCompletedAt))).FullWidth(),
				ui.HLine(),

				ui.HStack(ui.Text("nächster Lauf geplant"), ui.Spacer(), ui.Text(formatDate(status.NextPlannedAt))).FullWidth(),
				ui.HLine(),

				ui.HStack(ui.Text("letzter Fehler"), ui.Spacer(), ui.Text(lastErr(status.LastError))).FullWidth(),
				ui.HLine(),
			).Alignment(ui.Leading).FullWidth()).Footer(ui.SecondaryButton(func() {
			editSettingsPresented.Set(true)
		}).Title("Einstellungen bearbeiten")).Frame(ui.Frame{}.FullWidth()),
		ui.FixedSpacer(ui.L48, ui.L48),

		ui.IfFunc(len(status.Options.Actions) > 0, func() core.View {
			return ui.HStack(
				ui.ForEach(status.Options.Actions, func(t scheduler.CustomAction) core.View {
					return ui.SecondaryButton(func() {
						t.Action(wnd.Context())
					}).Title(t.Title)
				})...,
			).FullWidth().Gap(ui.L8).Alignment(ui.Trailing)
		}),

		ui.H2("Log-Einträge"),
		logView(wnd, sid, scheduleUseCases),
	).FullWidth().Alignment(ui.Leading)
}

func kindStr(kind scheduler.Kind) string {
	switch kind {
	case scheduler.Schedule:
		return "Wiederholt"
	case scheduler.OneShot:
		return "Einmalig"
	case scheduler.Manual:
		return "Manuell"
	default:
		return "Unknown"
	}
}

func kindIco(kind scheduler.Kind) core.SVG {
	switch kind {
	case scheduler.Schedule:
		return heroSolid.ArrowPathRoundedSquare
	case scheduler.OneShot:
		return heroSolid.RocketLaunch
	case scheduler.Manual:
		return heroSolid.Play
	default:
		return heroSolid.QuestionMarkCircle
	}
}

func stateStr(state scheduler.State) string {
	switch state {
	case scheduler.Stopped:
		return "beendet"
	case scheduler.Running:
		return "in Ausführung"
	case scheduler.Disabled:
		return "deaktiviert"
	case scheduler.Paused:
		return "pausiert"
	default:
		return "unknown"
	}
}

func formatDate(date time.Time) string {
	if date.IsZero() {
		return "undefiniert"
	}

	return date.Format(xtime.GermanDateTime)
}

func lastErr(err error) string {
	if err == nil {
		return "kein Fehler aufgetreten"
	}

	return err.Error()
}

func editSettingsDialog(wnd core.Window, status scheduler.StatusResult, editSettingsPresented *core.State[bool], scheduleUseCases scheduler.UseCases) core.View {
	return ui.Lazy(func() core.View {
		if !editSettingsPresented.Get() {
			return nil
		}

		state := core.AutoState[scheduler.Settings](wnd).Init(func() scheduler.Settings {
			tmp := status.Options.Defaults
			tmp.ID = status.Options.ID

			optSched, err := scheduleUseCases.FindSettingsByID(wnd.Subject(), status.Options.ID)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return tmp
			}

			if optSched.IsNone() {
				return tmp
			}

			return optSched.Unwrap()
		})

		return alert.Dialog(
			"Einstellungen bearbeiten",
			ui.Lazy(func() core.View {
				return ui.VStack(
					ui.HStack(ui.SecondaryButton(func() {
						if err := scheduleUseCases.DeleteSettingsByID(wnd.Subject(), status.Options.ID); err != nil {
							alert.ShowBannerError(wnd, err)
						}
						editSettingsPresented.Set(false)
					}).Title("Zurücksetzen")).Alignment(ui.Trailing).FullWidth(),
					form.Auto(form.AutoOptions{}, state).Frame(ui.Frame{}.FullWidth()),
				).FullWidth().Gap(ui.L8)
			}),
			editSettingsPresented,

			alert.Save(func() (close bool) {
				if err := scheduleUseCases.UpdateSettings(wnd.Subject(), state.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return false
				}

				return true
			}),
			alert.Cancel(nil),
		)

	})

}

func logView(wnd core.Window, id scheduler.ID, scheduleUseCases scheduler.UseCases) core.View {
	logs, err := xslices.Collect2(scheduleUseCases.ViewLogs(wnd.Subject(), id))
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.Table(
		ui.TableColumn(ui.Text("Zeit")),
		ui.TableColumn(ui.Text("Nachricht")),
		ui.TableColumn(ui.Text("weiteres")),
	).Rows(
		ui.ForEach(logs, func(t scheduler.LogEntry) ui.TTableRow {
			var more string
			if len(t.Values) > 0 {
				buf, _ := json.Marshal(t.Values)
				more = string(buf)
			}

			return ui.TableRow(
				ui.TableCell(ui.Text(t.Time.Format(time.RFC3339))),
				ui.TableCell(ui.Text(t.Msg)),
				ui.TableCell(ui.Text(more)),
			)
		})...,
	).Frame(ui.Frame{}.FullWidth())
}
