// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uischeduler

import (
	"errors"
	"fmt"
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/text/language"
	"log/slog"
	"maps"
	"slices"
	"time"

	"go.wdy.de/nago/application/scheduler"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/form"
)

func PageOverview(wnd core.Window, scheduleUseCases scheduler.UseCases) core.View {
	sid := scheduler.ID(wnd.Values()["id"])
	status, err := scheduleUseCases.Status(wnd.Subject(), sid)
	if err != nil {
		return alert.BannerError(err)
	}

	editSettingsPresented := core.AutoState[bool](wnd)
	selectedRun := core.AutoState[scheduler.RunID](wnd)
	logLevel := core.AutoState[slog.Level](wnd).Init(func() slog.Level { return slog.LevelDebug })
	logOffset := core.AutoState[int](wnd)
	logQuery := core.AutoState[string](wnd)

	runs, runsErr := xslices.Collect2(scheduleUseCases.ListRuns(wnd.Subject(), sid))

	running := status.State == scheduler.Running
	stopped := status.State == scheduler.Stopped

	showErrors := func(run scheduler.RunID) {
		selectedRun.Set(run)
		logLevel.Set(slog.LevelError)
		logOffset.Set(0)
	}

	return ui.VStack(
		ui.RedrawAtFixedRate[core.View](wnd, time.Second, nil),
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
						badge("", kindDetail(status)),
						badge("", string(status.Options.ID)),
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

		errorCallout(wnd, status, showErrors),
		ui.IfFunc(runsErr != nil, func() core.View { return alert.BannerError(runsErr) }),

		grid([5]int{1, 2, 4, 4, 4},
			nextRunKpi(wnd, status),
			kpi("Zuletzt gestartet", relTime(status.LastStartedAt), "", formatDate(wnd, status.LastStartedAt), ""),
			kpi("Zuletzt beendet", relTime(status.LastCompletedAt), "", lastRunHint(wnd, status), ""),
			failedKpi(status.Stats),
		),

		grid([5]int{1, 1, 2, 2, 2},
			card("Letzte Läufe", runsTable(wnd, runs, selectedRun, logOffset)).Frame(ui.Frame{}.FullWidth()),
			card("Konfiguration", settingsSummary(status), ui.SecondaryButton(func() {
				editSettingsPresented.Set(true)
			}).Title("Bearbeiten")).Frame(ui.Frame{}.FullWidth()),
		),

		card("Protokoll", logView(wnd, sid, scheduleUseCases, runs, selectedRun, logLevel, logOffset, logQuery)).Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Alignment(ui.TopLeading).Gap(ui.L16)
}

func statePill(state scheduler.State) core.View {
	switch state {
	case scheduler.Running:
		return badge(ui.SG0, "in Ausführung")
	case scheduler.Paused:
		return badge(ui.SV0, "wartet auf nächsten Lauf")
	case scheduler.Disabled:
		return badge(ui.SW0, "deaktiviert – nur manuelle Ausführung")
	default:
		return badge(ui.ST0, stateStr(state))
	}
}

// badge renders a label with the theme text color on the neutral background, so that the contrast is
// sufficient in light and dark mode. The semantic color is only used for the leading dot.
func badge(dot ui.Color, text string) core.View {
	return ui.HStack(
		ui.IfFunc(dot != "", func() core.View {
			return ui.VStack().
				BackgroundColor(dot).
				Border(ui.Border{}.Circle()).
				Frame(ui.Frame{}.Size(ui.L8, ui.L8))
		}),
		ui.Text(text).Font(ui.BodySmall).Color(ui.M8),
	).Gap(ui.L4).
		BackgroundColor(ui.M1).
		Border(ui.Border{}.Radius(ui.L16).Color(ui.M5).Width(ui.L1)).
		Padding(ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L2))
}

func kindDetail(status scheduler.StatusResult) string {
	s := status.Settings
	switch status.Options.Kind {
	case scheduler.Schedule:
		return "Wiederholt · alle " + formatDuration(max(s.PauseTime, time.Second))
	case scheduler.Cron:
		return fmt.Sprintf("Zeitgesteuert · %02d:%02d Uhr", s.CronHour, s.CronMinute)
	default:
		return kindStr(status.Options.Kind)
	}
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

func errorCallout(wnd core.Window, status scheduler.StatusResult, showErrors func(scheduler.RunID)) core.View {
	if status.LastRun.IsSome() {
		run := status.LastRun.Unwrap()
		if !run.Outcome.Failed() {
			return nil
		}

		return ui.HStack(
			ui.VStack(
				ui.Text("Letzter Lauf fehlgeschlagen · "+formatDate(wnd, run.StartedAt)).Font(ui.TitleSmall).Color(ui.SE0),
				ui.Text(run.Error),
			).Alignment(ui.Leading).Gap(ui.L4),
			ui.Spacer(),
			ui.IfFunc(run.LogFile != "", func() core.View {
				return ui.SecondaryButton(func() { showErrors(run.ID) }).Title("Im Protokoll anzeigen")
			}),
		).Gap(ui.L16).FullWidth().
			BackgroundColor(ui.ColorCardBody).
			Border(ui.Border{}.Radius(ui.L12).Color(ui.SE0).Width(ui.L1)).
			Padding(ui.Padding{}.All(ui.L16))
	}

	// backwards compatible fallback, e.g. for errors without run statistics
	if status.LastError == nil {
		return nil
	}

	return alert.Banner("Letzter Fehler", status.LastError.Error()).Frame(ui.Frame{}.FullWidth())
}

func nextRunKpi(wnd core.Window, status scheduler.StatusResult) core.View {
	switch {
	case status.State == scheduler.Running:
		return kpi("Nächster Lauf", "läuft gerade", ui.SG0, "seit "+relTime(status.LastStartedAt), "")
	case status.State == scheduler.Stopped || status.State == scheduler.Disabled || status.NextPlannedAt.IsZero():
		return kpi("Nächster Lauf", "–", "", "nicht geplant", "")
	default:
		d := time.Until(status.NextPlannedAt)
		return kpi("Nächster Lauf", "in "+countdown(d), "", formatDate(wnd, status.NextPlannedAt), "")
	}
}

func failedKpi(st scheduler.RunStats) core.View {
	if st.Total24h == 0 {
		return kpi("Fehler (24 h)", "–", "", "keine Läufe", "")
	}

	color := ui.Color("")
	if st.Failed24h > 0 {
		color = ui.SE0
	}

	ok := float64(st.Total24h-st.Failed24h) / float64(st.Total24h) * 100
	return kpi("Fehler (24 h)", fmt.Sprintf("%d von %d", st.Failed24h, st.Total24h), color, fmt.Sprintf("%.1f %% erfolgreich · Ø %s", ok, formatDuration(st.AvgDuration)), "")
}

func lastRunHint(wnd core.Window, status scheduler.StatusResult) string {
	hint := formatDate(wnd, status.LastCompletedAt)
	if status.LastRun.IsSome() && status.Stats.LastDuration > 0 {
		hint += " · Dauer " + formatDuration(status.Stats.LastDuration)
	}
	return hint
}

func outcomePill(o scheduler.Outcome) core.View {
	switch o {
	case scheduler.OutcomeRunning:
		return badge(ui.SV0, "läuft")
	case scheduler.OutcomeSucceeded:
		return badge(ui.SG0, "erfolgreich")
	case scheduler.OutcomeFailed:
		return badge(ui.SE0, "Fehler")
	case scheduler.OutcomePanicked:
		return badge(ui.SE0, "Absturz")
	case scheduler.OutcomeCanceled:
		return badge(ui.SW0, "abgebrochen")
	default:
		return badge(ui.ST0, "unbekannt")
	}
}

const maxRunRows = 10

func runsTable(wnd core.Window, runs []scheduler.Run, selected *core.State[scheduler.RunID], offset *core.State[int]) core.View {
	if len(runs) == 0 {
		return ui.Text("Noch keine Läufe erfasst.").Color(ui.ST0)
	}

	cur := selectedRunID(runs, selected.Get())
	shown := runs[:min(len(runs), maxRunRows)]

	return ui.VStack(
		ui.Table(
			ui.TableColumn(ui.Text("Start")),
			ui.TableColumn(ui.Text("Ergebnis")),
			ui.TableColumn(ui.Text("Dauer")),
			ui.TableColumn(ui.Text("Einträge")),
		).Rows(ui.ForEach(shown, func(r scheduler.Run) ui.TTableRow {
			row := ui.TableRow(
				ui.TableCell(ui.Text(shortDate(wnd, r.StartedAt))),
				ui.TableCell(outcomePill(r.Outcome)),
				ui.TableCell(ui.Text(formatDuration(r.Duration()))),
				ui.TableCell(ui.IfElse(r.LogFile != "" || r.Outcome == scheduler.OutcomeRunning, ui.Text(entriesStr(r)), ui.Text(entriesStr(r)+" · Protokoll gelöscht").Color(ui.ST0))),
			)

			if r.LogFile != "" || r.Outcome == scheduler.OutcomeRunning {
				row = row.Action(func() {
					selected.Set(r.ID)
					offset.Set(0)
				})
			}

			if r.ID == cur {
				row = row.BackgroundColor(ui.I1)
			}

			return row
		})...).Frame(ui.Frame{}.FullWidth()),
		ui.Text(fmt.Sprintf("%d Läufe in den letzten 30 Tagen, Protokolle der letzten %d Läufe werden aufbewahrt.", len(runs), scheduler.KeepRunLogs)).Font(ui.BodySmall).Color(ui.ST0),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}

func entriesStr(r scheduler.Run) string {
	s := fmt.Sprint(r.Entries)
	if r.Errors > 0 {
		s += fmt.Sprintf(" · %d Fehler", r.Errors)
	} else if r.Warnings > 0 {
		s += fmt.Sprintf(" · %d Warnungen", r.Warnings)
	}
	return s
}

// selectedRunID returns the explicitly selected run or the newest run. A selected run whose log has been
// deleted in the meantime (see [scheduler.KeepRunLogs]) falls back to the newest run.
func selectedRunID(runs []scheduler.Run, selected scheduler.RunID) scheduler.RunID {
	for _, r := range runs {
		if r.ID == selected && (r.LogFile != "" || r.Outcome == scheduler.OutcomeRunning) {
			return selected
		}
	}

	if len(runs) > 0 {
		return runs[0].ID
	}

	return ""
}

func settingsSummary(status scheduler.StatusResult) core.View {
	s, d := status.Settings, status.Options.Defaults
	custom := func(changed bool, v string) core.View {
		if status.CustomSettings && changed {
			return ui.HStack(ui.Text(v), badge(ui.I0, "angepasst")).Gap(ui.L8)
		}
		return ui.Text(v)
	}

	enabled := "aktiviert"
	if s.Disabled {
		enabled = "deaktiviert"
	}

	rows := []ui.TTableRow{
		kvRow("Automatische Ausführung", custom(s.Disabled != d.Disabled, enabled)),
		kvRow("Verzögerung nach Start", custom(s.StartDelay != d.StartDelay, formatDuration(s.StartDelay))),
	}

	switch status.Options.Kind {
	case scheduler.Schedule:
		rows = append(rows, kvRow("Pause zwischen Läufen", custom(s.PauseTime != d.PauseTime, formatDuration(s.PauseTime))))
	case scheduler.Cron:
		rows = append(rows,
			kvRow("Uhrzeit", custom(s.CronHour != d.CronHour || s.CronMinute != d.CronMinute, fmt.Sprintf("%02d:%02d Uhr", s.CronHour, s.CronMinute))),
			kvRow("Wiederholung", custom(s.PauseTime != d.PauseTime, formatDuration(s.PauseTime))),
		)
	}

	hint := "Alle Werte entsprechen dem Standard des Prozesses."
	if status.CustomSettings {
		hint = "Markierte Werte weichen vom Standard des Prozesses ab."
	}

	return ui.VStack(
		ui.Table(ui.TableColumn(nil), ui.TableColumn(nil)).Rows(rows...).
			CellPadding(ui.Padding{}.Vertical(ui.L8).Horizontal(ui.L4)).
			HeaderDividerColor("").
			Frame(ui.Frame{}.FullWidth()),
		ui.Text(hint).Font(ui.BodySmall).Color(ui.ST0),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}

func kvRow(key string, value core.View) ui.TTableRow {
	return ui.TableRow(
		ui.TableCell(ui.Text(key).Color(ui.ST0)).Alignment(ui.Leading),
		ui.TableCell(value).Alignment(ui.Trailing),
	)
}

func card(title string, body ...core.View) ui.DecoredView {
	return ui.VStack(
		ui.IfElse(title == "", nil, ui.Text(title).Font(ui.TitleMedium)),
	).Append(body...).
		Alignment(ui.TopLeading).
		Gap(ui.L12).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Padding(ui.Padding{}.All(ui.L16))
}

func kpi(title, value string, valueColor ui.Color, hint string, hintColor ui.Color) core.View {
	v := ui.Text(value).Font(ui.HeadlineMedium)
	if valueColor != "" {
		v = v.Color(valueColor)
	}

	h := ui.Text(hint).Font(ui.BodySmall).Color(ui.ST0)
	if hintColor != "" {
		h = h.Color(hintColor)
	}

	return card("", ui.Text(title).Font(ui.LabelLarge), v, h).Frame(ui.Frame{MinWidth: ui.L160}.FullWidth())
}

func grid(cols [5]int, children ...core.View) core.View {
	classes := []core.WindowSizeClass{core.SizeClassSmall, core.SizeClassMedium, core.SizeClassLarge, core.SizeClassXL, core.SizeClass2XL}
	l := cardlayout.Layout(children...)
	for i, c := range classes {
		l = l.Columns(c, cols[i])
	}

	return l.Frame(ui.Frame{}.FullWidth())
}

// shortDate omits the date for today.
func shortDate(wnd core.Window, t time.Time) string {
	t = t.In(wnd.Location())
	now := time.Now().In(wnd.Location())
	if t.YearDay() == now.YearDay() && t.Year() == now.Year() {
		return t.Format("15:04:05")
	}

	return t.Format("02.01. 15:04")
}

func relTime(t time.Time) string {
	if t.IsZero() {
		return "noch nie"
	}

	d := time.Since(t)
	if d < 0 {
		return "in " + countdown(-d)
	}

	return "vor " + countdown(d)
}

func countdown(d time.Duration) string {
	d = d.Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d s", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%d:%02d min", int(d.Minutes()), int(d.Seconds())%60)
	case d < 48*time.Hour:
		return fmt.Sprintf("%d:%02d h", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%d Tagen", int(d.Hours()/24))
	}
}

func formatDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "–"
	case d < time.Second:
		return fmt.Sprintf("%d ms", d.Milliseconds())
	case d < time.Minute:
		if d%time.Second == 0 {
			return fmt.Sprintf("%d s", int(d.Seconds()))
		}
		return fmt.Sprintf("%.1f s", d.Seconds())
	case d < time.Hour:
		if d%time.Minute == 0 {
			return fmt.Sprintf("%d min", int(d.Minutes()))
		}
		return fmt.Sprintf("%.1f min", d.Minutes())
	default:
		return fmt.Sprintf("%.1f h", d.Hours())
	}
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

func formatDate(wnd core.Window, t time.Time) string {
	if t.IsZero() {
		return "–"
	}

	return t.In(wnd.Location()).Format(xtime.GermanDateTime)
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

type logLevelOption struct {
	title string
	level slog.Level
}

var logLevels = []logLevelOption{
	{"Alle", slog.LevelDebug},
	{"Info", slog.LevelInfo},
	{"Warnung", slog.LevelWarn},
	{"Fehler", slog.LevelError},
}

func logView(wnd core.Window, id scheduler.ID, uc scheduler.UseCases, runs []scheduler.Run, selected *core.State[scheduler.RunID], level *core.State[slog.Level], offset *core.State[int], query *core.State[string]) core.View {
	runID := selectedRunID(runs, selected.Get())
	if runID == "" {
		return ui.Text("Noch keine Läufe erfasst.").Color(ui.ST0)
	}

	var run scheduler.Run
	for _, r := range runs {
		if r.ID == runID {
			run = r
		}
	}

	var levelButtons []core.View
	for _, l := range logLevels {
		action := func() {
			level.Set(l.level)
			offset.Set(0)
		}
		if l.level == level.Get() {
			levelButtons = append(levelButtons, ui.PrimaryButton(action).Title(l.title))
		} else {
			levelButtons = append(levelButtons, ui.SecondaryButton(action).Title(l.title))
		}
	}

	toolbar := ui.HStack(
		ui.Text("Lauf vom "+formatDate(wnd, run.StartedAt)).Font(ui.TitleSmall),
		ui.Spacer(),
		ui.HStack(levelButtons...).Gap(ui.L4),
		ui.TextField("", query.Get()).
			InputValue(query).
			Style(ui.TextFieldReduced).
			Leading(ui.ImageIcon(heroSolid.MagnifyingGlass)).
			Debounce(true),
	).Gap(ui.L8).Wrap(true).FullWidth()

	page, err := uc.ViewRunLog(wnd.Subject(), id, runID, scheduler.LogQuery{
		Offset:   offset.Get(),
		Limit:    scheduler.MaxLogPageSize,
		MinLevel: level.Get(),
		Text:     query.Get(),
	})

	if err != nil {
		return ui.VStack(toolbar, alert.BannerError(localizeRunLogErr(wnd, err))).Gap(ui.L12).FullWidth()
	}

	// the query changed and the offset is beyond the result
	if page.Total > 0 && page.Offset >= page.Total {
		offset.Set(0)
	}

	var body core.View
	if len(page.Entries) == 0 {
		body = ui.Text("Keine Einträge für diesen Filter.").Color(ui.ST0).Padding(ui.Padding{}.All(ui.L24))
	} else {
		body = ui.Table(
			ui.TableColumn(ui.Text("Zeit")).Width(ui.L160),
			ui.TableColumn(ui.Text("Stufe")).Width(ui.L80),
			ui.TableColumn(ui.Text("Nachricht")),
		).Rows(ui.ForEach(page.Entries, func(t scheduler.LogEntry) ui.TTableRow {
			return ui.TableRow(
				ui.TableCell(ui.Text(t.Time.In(wnd.Location()).Format("02.01.2006 15:04:05.000")).Color(ui.ST0)),
				ui.TableCell(levelPill(t.Level)),
				ui.TableCell(ui.VStack(
					ui.Text(t.Msg),
					ui.IfFunc(len(t.Values) > 0, func() core.View { return valueChips(t.Values) }),
				).Alignment(ui.Leading).Gap(ui.L4)),
			)
		})...).Frame(ui.Frame{}.FullWidth())
	}

	from, to := page.Offset+1, page.Offset+len(page.Entries)
	if len(page.Entries) == 0 {
		from = 0
	}

	pager := ui.HStack(
		ui.Text(fmt.Sprintf("%d–%d von %d Einträgen · neueste zuerst", from, to, page.Total)).Font(ui.BodySmall).Color(ui.ST0),
		ui.Spacer(),
		ui.SecondaryButton(func() { offset.Set(max(offset.Get()-scheduler.MaxLogPageSize, 0)) }).Title("Neuere").Enabled(page.Offset > 0),
		ui.SecondaryButton(func() { offset.Set(offset.Get() + scheduler.MaxLogPageSize) }).Title("Ältere").Enabled(to < page.Total),
	).Gap(ui.L8).FullWidth()

	return ui.VStack(toolbar, body, pager).Alignment(ui.Leading).Gap(ui.L12).FullWidth()
}

func levelPill(l slog.Level) core.View {
	switch {
	case l >= slog.LevelError:
		return badge(ui.SE0, "Fehler")
	case l >= slog.LevelWarn:
		return badge(ui.SW0, "Warnung")
	case l >= slog.LevelInfo:
		return badge(ui.ST0, "Info")
	default:
		return badge(ui.ST0, "Debug")
	}
}

func valueChips(values map[string]any) core.View {
	keys := slices.Sorted(maps.Keys(values))
	var chips []core.View
	for _, k := range keys {
		v := fmt.Sprint(values[k])
		if len(v) > 200 {
			v = v[:200] + "…"
		}
		chips = append(chips, ui.Text(k+"="+v).Font(ui.BodySmall).Color(ui.ST0))
	}

	return ui.HStack(chips...).Gap(ui.L8).Wrap(true).Alignment(ui.Leading)
}

var (
	strRunLogNotAvailable = i18n.MustString("nago.scheduler.run_log_not_available", i18n.Values{
		language.German:  "Protokoll nicht mehr verfügbar",
		language.English: "Log no longer available",
	})
	strRunLogNotAvailableMsgX = i18n.MustVarString("nago.scheduler.run_log_not_available_msg_x", i18n.Values{
		language.German:  "Das Protokoll dieses Laufs wurde bereits gelöscht. Es werden nur die Protokolle der letzten {amount} Läufe aufbewahrt.",
		language.English: "The log of this run has already been deleted. Only the logs of the last {amount} runs are kept.",
	})
	strRunNotFound = i18n.MustString("nago.scheduler.run_not_found", i18n.Values{
		language.German:  "Lauf nicht gefunden",
		language.English: "Run not found",
	})
	strRunNotFoundMsg = i18n.MustString("nago.scheduler.run_not_found_msg", i18n.Values{
		language.German:  "Der Lauf ist nicht mehr vorhanden. Laufstatistiken werden 30 Tage aufbewahrt.",
		language.English: "The run does not exist anymore. Run statistics are kept for 30 days.",
	})
)

// localizeRunLogErr turns the expected domain errors of [scheduler.ViewRunLog] into localized errors.
func localizeRunLogErr(wnd core.Window, err error) error {
	switch {
	case errors.Is(err, scheduler.ErrRunLogNotAvailable):
		return std.NewLocalizedError(
			strRunLogNotAvailable.Get(wnd),
			strRunLogNotAvailableMsgX.Get(wnd, i18n.Int("amount", scheduler.KeepRunLogs)),
		).WithError(err)
	case errors.Is(err, scheduler.ErrRunNotFound):
		return std.NewLocalizedError(strRunNotFound.Get(wnd), strRunNotFoundMsg.Get(wnd)).WithError(err)
	default:
		return err
	}
}
