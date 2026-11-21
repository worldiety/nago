// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uischeduler

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/scheduler"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/tags"
	"golang.org/x/text/language"
)

// DashboardModel provides the data of a [Dashboard]. The embedder checks its own rights in these functions.
type DashboardModel struct {
	// Status returns the state and the statistics of the scheduler, see [scheduler.Status].
	Status func() (scheduler.StatusResult, error)
	// Runs returns a page of the runs, newest first, and the amount of all runs, see [scheduler.Manager.RunsPage].
	Runs func(offset, limit int) ([]scheduler.Run, int, error)
	// RunLog returns a page of the log of a run, see [scheduler.ViewRunLog].
	RunLog func(run scheduler.RunID, q scheduler.LogQuery) (scheduler.LogPage, error)
	// Location is the zone of all times. Nil means the zone of the window.
	Location *time.Location
	// Config is the content of the card "Konfiguration".
	Config core.View
	// Hints are shown above the statistics, after the error of the last failed run.
	Hints []core.View
}

// runsPageSize is the amount of runs in a page of "Letzte Läufe".
const runsPageSize = 10

// Dashboard shows the statistics, the latest runs, the configuration and the log of a run of a scheduler, without a
// head, breadcrumbs, buttons of the scheduler and [ui.H1], so that the embedder sets its own head and actions. It
// redraws itself each second and reads the runs by pages.
func Dashboard(wnd core.Window, model DashboardModel) core.View {
	loc := model.Location
	if loc == nil {
		loc = wnd.Location()
	}

	selectedRun := core.AutoState[scheduler.RunID](wnd)
	selectedStart := core.AutoState[time.Time](wnd)
	logLevel := core.AutoState[slog.Level](wnd).Init(func() slog.Level { return slog.LevelDebug })
	logOffset := core.AutoState[int](wnd)
	logQuery := core.AutoState[string](wnd)
	runsOffset := core.AutoState[int](wnd)

	redraw := ui.RedrawAtFixedRate[core.View](wnd, time.Second, nil)

	if model.Status == nil {
		return alert.BannerError(errors.New("dashboard without status"))
	}

	status, err := model.Status()
	if err != nil {
		return ui.VStack(redraw, alert.BannerError(err)).FullWidth()
	}

	var runs []scheduler.Run
	var total int
	var runsErr error
	if model.Runs != nil {
		runs, total, runsErr = model.Runs(runsOffset.Get(), runsPageSize)
		if runsErr == nil && runsOffset.Get() > 0 && runsOffset.Get() >= total {
			// the runs have been deleted meanwhile
			runsOffset.Set(0)
			runs, total, runsErr = model.Runs(0, runsPageSize)
		}
	}

	selectRun := func(run scheduler.Run) {
		selectedRun.Set(run.ID)
		selectedStart.Set(run.StartedAt)
		logOffset.Set(0)
	}

	showErrors := func(run scheduler.Run) {
		selectRun(run)
		logLevel.Set(slog.LevelError)
	}

	// without a selection, the newest run
	selected := scheduler.Run{ID: selectedRun.Get(), StartedAt: selectedStart.Get()}
	if selected.ID == "" && status.LastRun.IsSome() {
		selected = status.LastRun.Unwrap()
	}

	keep := keepLogs(status)

	views := []core.View{redraw, errorCallout(loc, status, showErrors)}
	views = append(views, model.Hints...)
	views = append(views,
		ui.IfFunc(runsErr != nil, func() core.View { return alert.BannerError(runsErr) }),

		grid([5]int{1, 2, 4, 4, 4},
			nextRunKpi(loc, status),
			kpi("Zuletzt gestartet", relTime(status.LastStartedAt), "", formatDate(loc, status.LastStartedAt), ""),
			kpi("Zuletzt beendet", relTime(status.LastCompletedAt), "", lastRunHint(loc, status), ""),
			failedKpi(status.Stats),
		),

		grid([5]int{1, 1, 2, 2, 2},
			card("Letzte Läufe", runsTable(loc, runs, total, keep, selected.ID, selectRun, runsOffset)).Frame(ui.Frame{}.FullWidth()),
			card("Konfiguration", model.Config).Frame(ui.Frame{}.FullWidth()),
		),

		card("Protokoll", logView(wnd, loc, model, selected, keep, logLevel, logOffset, logQuery)).Frame(ui.Frame{}.FullWidth()),
	)

	return ui.VStack(views...).FullWidth().Alignment(ui.TopLeading).Gap(ui.L16)
}

// keepLogs returns the amount of runs, whose log is kept.
func keepLogs(status scheduler.StatusResult) int {
	if status.Settings.KeepRunLogs > 0 {
		return status.Settings.KeepRunLogs
	}

	return scheduler.KeepRunLogs
}

func errorCallout(loc *time.Location, status scheduler.StatusResult, showErrors func(scheduler.Run)) core.View {
	if status.LastRun.IsSome() {
		run := status.LastRun.Unwrap()
		if !run.Outcome.Failed() {
			return nil
		}

		return ui.HStack(
			ui.VStack(
				ui.Text("Letzter Lauf fehlgeschlagen · "+formatDate(loc, run.StartedAt)).Font(ui.TitleSmall).Color(ui.SE0),
				ui.Text(run.Error),
			).Alignment(ui.Leading).Gap(ui.L4),
			ui.Spacer(),
			ui.IfFunc(run.LogFile != "", func() core.View {
				return ui.SecondaryButton(func() { showErrors(run) }).Title("Im Protokoll anzeigen")
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

func nextRunKpi(loc *time.Location, status scheduler.StatusResult) core.View {
	switch {
	case status.State == scheduler.Running:
		return kpi("Nächster Lauf", "läuft gerade", ui.SG0, "seit "+relTime(status.LastStartedAt), "")
	case status.Suspended:
		return kpi("Nächster Lauf", "pausiert", ui.SW0, "bis zum Fortsetzen", "")
	case status.State == scheduler.Stopped || status.State == scheduler.Disabled || status.NextPlannedAt.IsZero():
		return kpi("Nächster Lauf", "–", "", "nicht geplant", "")
	default:
		d := time.Until(status.NextPlannedAt)
		return kpi("Nächster Lauf", "in "+countdown(d), "", formatDate(loc, status.NextPlannedAt), "")
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

func lastRunHint(loc *time.Location, status scheduler.StatusResult) string {
	hint := formatDate(loc, status.LastCompletedAt)
	if status.LastRun.IsSome() && status.Stats.LastDuration > 0 {
		hint += " · Dauer " + formatDuration(status.Stats.LastDuration)
	}

	if status.LastRun.IsSome() {
		if run := status.LastRun.Unwrap(); run.Outcome != scheduler.OutcomeRunning {
			hint += " · " + outcomeStr(run)
		}
	}

	return hint
}

// outcomeStr returns the result of a run, e.g. "abgebrochen: Neustart".
func outcomeStr(r scheduler.Run) string {
	switch r.Outcome {
	case scheduler.OutcomeRunning:
		return "läuft"
	case scheduler.OutcomeSucceeded:
		return "erfolgreich"
	case scheduler.OutcomeFailed:
		return "Fehler"
	case scheduler.OutcomePanicked:
		return "Absturz"
	case scheduler.OutcomeCanceled:
		if r.CanceledBy != "" {
			return "abgebrochen: " + r.CanceledBy
		}
		return "abgebrochen"
	default:
		return "unbekannt"
	}
}

func outcomePill(r scheduler.Run) core.View {
	switch r.Outcome {
	case scheduler.OutcomeRunning:
		return tags.StatusBadge(ui.SV0, outcomeStr(r))
	case scheduler.OutcomeSucceeded:
		return tags.StatusBadge(ui.SG0, outcomeStr(r))
	case scheduler.OutcomeFailed, scheduler.OutcomePanicked:
		return tags.StatusBadge(ui.SE0, outcomeStr(r))
	case scheduler.OutcomeCanceled:
		return tags.StatusBadge(ui.SW0, outcomeStr(r))
	default:
		return tags.StatusBadge(ui.ST0, outcomeStr(r))
	}
}

// triggerStr returns why a run started, e.g. "von Hand: Anna".
func triggerStr(r scheduler.Run) string {
	switch r.Trigger {
	case scheduler.TriggerCatchUp:
		return "nachgeholt"
	case scheduler.TriggerManual:
		if r.By != "" {
			return "von Hand: " + r.By
		}
		return "von Hand"
	default:
		return "Plan"
	}
}

func hasLog(r scheduler.Run) bool {
	return r.LogFile != "" || r.Outcome == scheduler.OutcomeRunning
}

func runsTable(loc *time.Location, runs []scheduler.Run, total, keep int, selected scheduler.RunID, onSelect func(scheduler.Run), offset *core.State[int]) core.View {
	if total == 0 {
		return ui.Text("Noch keine Läufe erfasst.").Color(ui.ST0)
	}

	from, to := offset.Get()+1, offset.Get()+len(runs)
	if len(runs) == 0 {
		from = 0
	}

	return ui.VStack(
		ui.Table(
			ui.TableColumn(ui.Text("Start")),
			ui.TableColumn(ui.Text("Ergebnis")),
			ui.TableColumn(ui.Text("Dauer")),
			ui.TableColumn(ui.Text("Einträge")),
		).Rows(ui.ForEach(runs, func(r scheduler.Run) ui.TTableRow {
			row := ui.TableRow(
				// the number and the trigger below the start keep the columns of the admin center as before
				ui.TableCell(ui.VStack(
					ui.Text(shortDate(loc, r.StartedAt)),
					ui.Text(fmt.Sprintf("Nr. %d · %s", r.Nr, triggerStr(r))).Font(ui.BodySmall).Color(ui.ST0),
				).Alignment(ui.Leading).Gap(ui.L2)),
				ui.TableCell(outcomePill(r)),
				ui.TableCell(ui.Text(formatDuration(r.Duration()))),
				ui.TableCell(ui.IfElse(hasLog(r), ui.Text(entriesStr(r)), ui.Text(entriesStr(r)+" · Protokoll gelöscht").Color(ui.ST0))),
			).Action(func() { onSelect(r) })

			if r.ID == selected {
				row = row.BackgroundColor(ui.I1)
			}

			return row
		})...).Frame(ui.Frame{}.FullWidth()),
		ui.IfFunc(total > runsPageSize, func() core.View {
			return ui.HStack(
				ui.Text(fmt.Sprintf("%d–%d von %d Läufen · neueste zuerst", from, to, total)).Font(ui.BodySmall).Color(ui.ST0),
				ui.Spacer(),
				ui.SecondaryButton(func() { offset.Set(max(offset.Get()-runsPageSize, 0)) }).Title("Neuere").Enabled(offset.Get() > 0),
				ui.SecondaryButton(func() { offset.Set(offset.Get() + runsPageSize) }).Title("Ältere").Enabled(to < total),
			).Gap(ui.L8).FullWidth()
		}),
		ui.Text(fmt.Sprintf("%d Läufe in den letzten 30 Tagen, Protokolle der letzten %d Läufe werden aufbewahrt.", total, keep)).Font(ui.BodySmall).Color(ui.ST0),
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
func shortDate(loc *time.Location, t time.Time) string {
	t = t.In(loc)
	now := time.Now().In(loc)
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

func formatDate(loc *time.Location, t time.Time) string {
	if t.IsZero() {
		return "–"
	}

	return t.In(loc).Format(xtime.GermanDateTime)
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

func logView(wnd core.Window, loc *time.Location, model DashboardModel, run scheduler.Run, keep int, level *core.State[slog.Level], offset *core.State[int], query *core.State[string]) core.View {
	if run.ID == "" || model.RunLog == nil {
		return ui.Text("Noch keine Läufe erfasst.").Color(ui.ST0)
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
		ui.Text("Lauf vom "+formatDate(loc, run.StartedAt)).Font(ui.TitleSmall),
		ui.Spacer(),
		ui.HStack(levelButtons...).Gap(ui.L4),
		ui.TextField("", query.Get()).
			InputValue(query).
			Style(ui.TextFieldReduced).
			Leading(ui.ImageIcon(heroSolid.MagnifyingGlass)).
			Debounce(true),
	).Gap(ui.L8).Wrap(true).FullWidth()

	page, err := model.RunLog(run.ID, scheduler.LogQuery{
		Offset:   offset.Get(),
		Limit:    scheduler.MaxLogPageSize,
		MinLevel: level.Get(),
		Text:     query.Get(),
	})

	if err != nil {
		return ui.VStack(toolbar, runLogErr(wnd, err, keep)).Gap(ui.L12).FullWidth()
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
				ui.TableCell(ui.Text(t.Time.In(loc).Format("02.01.2006 15:04:05.000")).Color(ui.ST0)),
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
		return tags.StatusBadge(ui.SE0, "Fehler")
	case l >= slog.LevelWarn:
		return tags.StatusBadge(ui.SW0, "Warnung")
	case l >= slog.LevelInfo:
		return tags.StatusBadge(ui.ST0, "Info")
	default:
		return tags.StatusBadge(ui.ST0, "Debug")
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

// runLogErr shows the expected domain errors of [scheduler.ViewRunLog] as a hint, without logging them on each
// redraw, and every other error as an error banner.
func runLogErr(wnd core.Window, err error, keep int) core.View {
	switch {
	case errors.Is(err, scheduler.ErrRunLogNotAvailable):
		return alert.Banner(strRunLogNotAvailable.Get(wnd), strRunLogNotAvailableMsgX.Get(wnd, i18n.Int("amount", keep)))
	case errors.Is(err, scheduler.ErrRunNotFound):
		return alert.Banner(strRunNotFound.Get(wnd), strRunNotFoundMsg.Get(wnd))
	default:
		return alert.BannerError(err)
	}
}
