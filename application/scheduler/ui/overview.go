// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uischeduler

import (
	"fmt"
	"time"

	"go.wdy.de/nago/application/scheduler"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/tags"
)

func PageOverview(wnd core.Window, scheduleUseCases scheduler.UseCases) core.View {
	sid := scheduler.ID(wnd.Values()["id"])
	status, err := scheduleUseCases.Status(wnd.Subject(), sid)
	if err != nil {
		return alert.BannerError(err)
	}

	editSettingsPresented := core.AutoState[bool](wnd)

	running := status.State == scheduler.Running
	stopped := status.State == scheduler.Stopped

	return ui.VStack(
		editSettingsDialog(wnd, status, editSettingsPresented, scheduleUseCases),
		breadcrumb.Breadcrumbs().
			Item("Admin Center", func() { wnd.Navigation().BackwardTo("admin", nil) }).
			Item("Hintergrundprozesse", func() { wnd.Navigation().BackwardTo("admin", nil) }).
			ClampLeading(),

		// header: identity, state and context dependent actions
		ui.HStack(
			ui.HStack(
				ui.VStack(
					ui.H1(status.Options.Name),
					ui.Text(status.Options.Description).Color(ui.ST0),
					ui.HStack(
						statePill(status.State),
						tags.StatusBadge("", kindDetail(status)),
						tags.StatusBadge("", string(status.Options.ID)),
					).Gap(ui.L8).Wrap(true),
				).Alignment(ui.Leading).Gap(ui.L4),
			).Alignment(ui.TopLeading).Gap(ui.L16),
			ui.Spacer(),
			ui.HStack(
				ui.PrimaryButton(func() {
					go func() {
						if err := scheduleUseCases.ExecuteNow(wnd.Subject(), sid); err != nil {
							alert.ShowBannerError(wnd, err)
						}
					}()
				}).Title("Jetzt ausführen").PreIcon(heroSolid.Play).Enabled(!running),
				ui.IfElse(stopped,
					ui.SecondaryButton(func() {
						if err := scheduleUseCases.Start(wnd.Subject(), sid); err != nil {
							alert.ShowBannerError(wnd, err)
						}
					}).Title("Starten"),
					ui.SecondaryButton(func() {
						if err := scheduleUseCases.Stop(wnd.Subject(), sid); err != nil {
							alert.ShowBannerError(wnd, err)
						}
					}).Title("Beenden"),
				),
				moreMenu(wnd, status, editSettingsPresented),
			).Gap(ui.L8),
		).Alignment(ui.TopLeading).Wrap(true).Gap(ui.L16).FullWidth(),

		Dashboard(wnd, DashboardModel{
			Status: func() (scheduler.StatusResult, error) {
				return status, nil
			},
			Runs: func(offset, limit int) ([]scheduler.Run, int, error) {
				return listRunsPage(wnd, scheduleUseCases, sid, offset, limit)
			},
			RunLog: func(run scheduler.RunID, q scheduler.LogQuery) (scheduler.LogPage, error) {
				return scheduleUseCases.ViewRunLog(wnd.Subject(), sid, run, q)
			},
			// the zone of the scheduler, which its cron time uses as well
			Location: schedulerLocation(status.Options),
			Config: ui.VStack(
				settingsSummary(status, isSmall(wnd)),
				ui.SecondaryButton(func() {
					editSettingsPresented.Set(true)
				}).Title("Bearbeiten"),
			).Alignment(ui.TopLeading).Gap(ui.L12).FullWidth(),
		}),
	).FullWidth().Alignment(ui.TopLeading).Gap(ui.L16)
}

// listRunsPage reads a page of the runs, also with use cases which do not provide [scheduler.ListRunsPage].
func listRunsPage(wnd core.Window, uc scheduler.UseCases, id scheduler.ID, offset, limit int) ([]scheduler.Run, int, error) {
	if uc.ListRunsPage != nil {
		return uc.ListRunsPage(wnd.Subject(), id, offset, limit)
	}

	runs, err := xslices.Collect2(uc.ListRuns(wnd.Subject(), id))
	if err != nil {
		return nil, 0, err
	}

	from := min(max(offset, 0), len(runs))
	return runs[from:min(from+max(limit, 0), len(runs))], len(runs), nil
}

func statePill(state scheduler.State) core.View {
	switch state {
	case scheduler.Running:
		return tags.StatusBadge(ui.SG0, "in Ausführung")
	case scheduler.Paused:
		return tags.StatusBadge(ui.SV0, "wartet auf nächsten Lauf")
	case scheduler.Disabled:
		return tags.StatusBadge(ui.SW0, "deaktiviert – nur manuelle Ausführung")
	default:
		return tags.StatusBadge(ui.ST0, stateStr(state))
	}
}

func kindDetail(status scheduler.StatusResult) string {
	s := status.Settings
	switch status.Options.Kind {
	case scheduler.Schedule:
		return "Wiederholt · alle " + formatDuration(max(s.PauseTime, time.Second))
	case scheduler.Cron:
		return fmt.Sprintf("Zeitgesteuert · %02d:%02d Uhr %s", s.CronHour, s.CronMinute, zoneName(schedulerLocation(status.Options)))
	default:
		return kindStr(status.Options.Kind)
	}
}

// schedulerLocation returns the zone of the cron time of a scheduler: its location or [time.Local].
func schedulerLocation(opts scheduler.Options) *time.Location {
	if opts.Location == nil {
		return time.Local
	}

	return opts.Location
}

// zoneName names a zone, e.g. Europe/Berlin, or for the zone of the server its abbreviation, e.g. UTC.
func zoneName(loc *time.Location) string {
	if loc == time.Local || loc.String() == "Local" {
		name, _ := time.Now().In(loc).Zone()
		return name
	}

	return loc.String()
}

func moreMenu(wnd core.Window, status scheduler.StatusResult, editSettingsPresented *core.State[bool]) core.View {
	var groups []ui.TMenuGroup
	if len(status.Options.Actions) > 0 {
		var items []ui.TMenuItem
		for _, a := range status.Options.Actions {
			items = append(items, ui.MenuItem(func() { a.Action(wnd.Context()) }, ui.Text(a.Title)))
		}
		groups = append(groups, ui.MenuGroup(items...))
	}

	groups = append(groups, ui.MenuGroup(
		ui.MenuItem(func() { editSettingsPresented.Set(true) }, ui.Text("Einstellungen bearbeiten")),
	))

	return ui.Menu(ui.SecondaryButton(nil).PreIcon(heroSolid.EllipsisHorizontal).AccessibilityLabel("Weitere Aktionen"), groups...)
}

func settingsSummary(status scheduler.StatusResult, small bool) core.View {
	s, d := status.Settings, status.Options.Defaults
	custom := func(changed bool, v string) core.View {
		if status.CustomSettings && changed {
			return ui.HStack(ui.Text(v), tags.StatusBadge(ui.I0, "angepasst")).Gap(ui.L8)
		}
		return ui.Text(v)
	}

	enabled := "aktiviert"
	if s.Disabled {
		enabled = "deaktiviert"
	}

	rows := []kv{
		kvRow("Automatische Ausführung", custom(s.Disabled != d.Disabled, enabled)),
		kvRow("Verzögerung nach Start", custom(s.StartDelay != d.StartDelay, formatDuration(s.StartDelay))),
	}

	switch status.Options.Kind {
	case scheduler.Schedule:
		rows = append(rows, kvRow("Pause zwischen Läufen", custom(s.PauseTime != d.PauseTime, formatDuration(s.PauseTime))))
	case scheduler.Cron:
		rows = append(rows,
			kvRow("Uhrzeit", custom(s.CronHour != d.CronHour || s.CronMinute != d.CronMinute, fmt.Sprintf("%02d:%02d Uhr %s", s.CronHour, s.CronMinute, zoneName(schedulerLocation(status.Options))))),
			kvRow("Wiederholung", custom(s.PauseTime != d.PauseTime, formatDuration(s.PauseTime))),
		)
	}

	hint := "Alle Werte entsprechen dem Standard des Prozesses."
	if status.CustomSettings {
		hint = "Markierte Werte weichen vom Standard des Prozesses ab."
	}

	var list core.View
	if small {
		// below 768 px each label stands above its value
		fields := make([]core.View, 0, len(rows))
		for _, r := range rows {
			fields = append(fields, divided(ui.VStack(field(r.key, r.value)).Alignment(ui.Leading).FullWidth().WithPadding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L4))))
		}
		list = ui.VStack(fields...).Alignment(ui.Leading).FullWidth()
	} else {
		tableRows := make([]ui.TTableRow, 0, len(rows))
		for _, r := range rows {
			tableRows = append(tableRows, ui.TableRow(
				ui.TableCell(ui.Text(r.key).Color(ui.ST0)).Alignment(ui.Leading),
				ui.TableCell(r.value).Alignment(ui.Trailing),
			))
		}
		list = ui.Table(ui.TableColumn(nil), ui.TableColumn(nil)).Rows(tableRows...).
			CellPadding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L4)).
			HeaderDividerColor("").
			Frame(ui.Frame{}.FullWidth())
	}

	return ui.VStack(
		list,
		ui.Text(hint).Font(ui.BodySmall).Color(ui.ST0),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}

// kv is a label and its value in the summary of the settings.
type kv struct {
	key   string
	value core.View
}

func kvRow(key string, value core.View) kv {
	return kv{key: key, value: value}
}

func kindStr(kind scheduler.Kind) string {
	switch kind {
	case scheduler.Schedule:
		return "Wiederholt"
	case scheduler.OneShot:
		return "Einmalig"
	case scheduler.Manual:
		return "Manuell"
	case scheduler.Cron:
		return "Zeitgesteuert"
	default:
		return "Unknown"
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
