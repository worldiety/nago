// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimail

import (
	"fmt"
	"strings"
	"time"

	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/barchart"
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/piechart"
	"go.wdy.de/nago/presentation/ui/tags"
)

type timeRange struct {
	key    string
	title  string
	dur    time.Duration
	bucket time.Duration // chart resolution
}

var timeRanges = []timeRange{
	{"24h", "24 Stunden", 24 * time.Hour, time.Hour},
	{"7d", "7 Tage", 7 * 24 * time.Hour, 24 * time.Hour},
	{"30d", "30 Tage", 30 * 24 * time.Hour, 24 * time.Hour},
}

func findTimeRange(key string) timeRange {
	for _, r := range timeRanges {
		if r.key == key {
			return r
		}
	}

	return timeRanges[1]
}

// DashboardPage shows the health and statistics of the mail system.
func DashboardPage(wnd core.Window, pages Pages, uc mail.UseCases) core.View {
	tr := findTimeRange(wnd.Values()["range"])

	now := time.Now()
	stats, err := uc.Statistics(wnd.Subject(), mail.StatisticsOptions{From: now.Add(-tr.dur), To: now})
	if err != nil {
		return alert.BannerError(err)
	}

	var rangeButtons []core.View
	for _, r := range timeRanges {
		action := func() { wnd.Navigation().ForwardTo(pages.Dashboard, core.Values{"range": r.key}) }
		if r.key == tr.key {
			rangeButtons = append(rangeButtons, ui.PrimaryButton(action).Title(r.title))
		} else {
			rangeButtons = append(rangeButtons, ui.SecondaryButton(action).Title(r.title))
		}
	}

	return ui.VStack(
		header(wnd, pages, tabDashboard, "E-Mail Übersicht", nil, ui.HStack(rangeButtons...).Gap(ui.L4)),
		problemBanner(wnd, pages, stats),
		ui.IfFunc(!stats.HasStats, func() core.View {
			return alert.Banner("Noch keine Statistiken", "Statistiken werden ab jetzt bei jedem Versandversuch erfasst. Die Werte der Warteschlange sind bereits verfügbar.").Intent(alert.IntentOk).Frame(ui.Frame{}.FullWidth())
		}),
		kpiRow(wnd, pages, stats),
		grid([5]int{1, 1, 2, 2, 2},
			card("Durchsatz", throughputChart(wnd, stats, tr)).Frame(ui.Frame{}.FullWidth()),
			card("Status der Warteschlange", statusDonut(stats)).Frame(ui.Frame{}.FullWidth()),
		),
		grid([5]int{1, 1, 2, 2, 2},
			card("Was ist gerade kaputt?", problemList(wnd, pages, stats)).Frame(ui.Frame{}.FullWidth()),
			card("Letzte Fehler", recentFailures(wnd, pages, stats)).Frame(ui.Frame{}.FullWidth()),
		),
	).Alignment(ui.TopLeading).Gap(ui.L16).FullWidth()
}

type problem struct {
	severity ui.Color
	title    string
	detail   string
	action   string
	fn       func()
}

func findProblems(wnd core.Window, pages Pages, stats mail.StatisticsResult) []problem {
	var res []problem
	if !stats.Scheduler.Running {
		res = append(res, problem{severity: ui.SE0, title: "Scheduler läuft nicht", detail: "Mails werden aktuell nicht verarbeitet."})
	}

	if stats.Scheduler.NoSmtpServer || (len(stats.Servers) == 0 && stats.Queue.Queued+stats.Queue.Error > 0) {
		res = append(res, problem{severity: ui.SE0, title: "Kein SMTP-Server konfiguriert", detail: "Es ist kein SMTP-Server mit der Systemgruppe geteilt.", action: "SMTP-Server", fn: func() {
			wnd.Navigation().ForwardTo(pages.SmtpServers, nil)
		}})
	}

	for _, srv := range stats.Servers {
		if srv.Health.ConsecutiveFailures > 0 {
			sev := ui.SW0
			if srv.Health.ConsecutiveFailures >= 3 {
				sev = ui.SE0
			}
			res = append(res, problem{
				severity: sev,
				title:    fmt.Sprintf("SMTP „%s“ – %d Fehler in Folge", srv.Name, srv.Health.ConsecutiveFailures),
				detail:   srv.Health.LastError,
				action:   "Server öffnen",
				fn:       func() { wnd.Navigation().ForwardTo(pages.SmtpServers, nil) },
			})
		}
	}

	if stats.Queue.Stuck > 0 {
		res = append(res, problem{
			severity: ui.SW0,
			title:    fmt.Sprintf("%d Mails hängen", stats.Queue.Stuck),
			detail:   fmt.Sprintf("Nicht versendet seit mehr als %s.", formatDuration(stats.Options.StuckAfter)),
			action:   "Anzeigen",
			fn:       func() { wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, core.Values{"stuck": "1"}) },
		})
	}

	if stats.Queue.Suppressed > 0 {
		res = append(res, problem{
			severity: ui.SE0,
			title:    fmt.Sprintf("%d Mails vom Spam-Schutz unterdrückt", stats.Queue.Suppressed),
			detail:   "Dieselbe Mail wurde auffällig oft an denselben Empfänger versendet. Das deutet auf eine Endlosschleife in der Anwendung hin.",
			action:   "Anzeigen",
			fn: func() {
				wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, core.Values{"status": string(mail.StatusSuppressed)})
			},
		})
	}

	for _, name := range stats.Scheduler.RateLimited {
		res = append(res, problem{
			severity: ui.SW0,
			title:    fmt.Sprintf("SMTP „%s“ – Ratenbegrenzung erreicht", name),
			detail:   "Weitere Mails werden verzögert, bis wieder Kontingent verfügbar ist.",
			action:   "Server öffnen",
			fn:       func() { wnd.Navigation().ForwardTo(pages.SmtpServers, nil) },
		})
	}

	if stats.Queue.Failed > 0 {
		res = append(res, problem{
			severity: ui.SE0,
			title:    fmt.Sprintf("%d Mails endgültig fehlgeschlagen", stats.Queue.Failed),
			detail:   "Diese Mails werden nicht mehr automatisch wiederholt.",
			action:   "Anzeigen",
			fn: func() {
				wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, core.Values{"status": string(mail.StatusFailed)})
			},
		})
	}

	if len(stats.TopErrors) > 0 && stats.TopErrors[0].Count > 1 {
		e := stats.TopErrors[0]
		res = append(res, problem{
			severity: ui.SW0,
			title:    fmt.Sprintf("Häufigster Fehler (%d×)", e.Count),
			detail:   e.Message,
			action:   "Filtern",
			fn: func() {
				wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, core.Values{"status": string(mail.StatusError) + "," + string(mail.StatusFailed) + "," + string(mail.StatusSuppressed)})
			},
		})
	}

	return res
}

func problemBanner(wnd core.Window, pages Pages, stats mail.StatisticsResult) core.View {
	problems := findProblems(wnd, pages, stats)
	if len(problems) == 0 {
		return alert.Banner("Alles in Ordnung", "Es sind aktuell keine Probleme bekannt.").Intent(alert.IntentSuccess).Frame(ui.Frame{}.FullWidth())
	}

	var titles []string
	for _, p := range problems {
		titles = append(titles, p.title)
	}

	intent := alert.IntentWarning
	for _, p := range problems {
		if p.severity == ui.SE0 {
			intent = alert.IntentError
		}
	}

	return alert.Banner(fmt.Sprintf("%d Probleme erkannt", len(problems)), strings.Join(titles, " · ")).Intent(intent).Frame(ui.Frame{}.FullWidth())
}

func problemList(wnd core.Window, pages Pages, stats mail.StatisticsResult) core.View {
	problems := findProblems(wnd, pages, stats)
	if len(problems) == 0 {
		return ui.Text("Keine Probleme erkannt.").Color(ui.SG0)
	}

	var rows []core.View
	for i, p := range problems {
		if i > 0 {
			rows = append(rows, ui.HLine().Padding(ui.Padding{}))
		}
		rows = append(rows, ui.HStack(
			ui.VStack().BackgroundColor(p.severity).Frame(ui.Frame{Width: ui.L12, Height: ui.L12}).Border(ui.Border{}.Circle()),
			ui.VStack(
				ui.Text(p.title).Font(ui.TitleSmall),
				ui.IfElse(p.detail == "", nil, ui.Text(p.detail).Font(ui.MonoSmall).Color(ui.ST0)),
			).Alignment(ui.Leading).Gap(ui.L4),
			ui.Spacer(),
			ui.IfElse(p.fn == nil, nil, ui.SecondaryButton(p.fn).Title(p.action)),
		).Gap(ui.L12).FullWidth())
	}

	return ui.VStack(rows...).Gap(ui.L8).FullWidth()
}

func trend(cur, prev int, higherIsBetter bool) (string, ui.Color) {
	if prev == 0 {
		if cur == 0 {
			return "unverändert", ""
		}
		return "kein Vergleichswert", ""
	}

	diff := float64(cur-prev) / float64(prev) * 100
	arrow := "▲"
	if diff < 0 {
		arrow = "▼"
	}

	good := (diff >= 0) == higherIsBetter
	col := ui.SE0
	if good {
		col = ui.SG0
	}

	if diff == 0 {
		return "unverändert", ""
	}

	return fmt.Sprintf("%s %.0f %% ggü. Vorzeitraum", arrow, diff), col
}

func kpiRow(wnd core.Window, pages Pages, stats mail.StatisticsResult) core.View {
	t := stats.Totals
	p := stats.Previous

	sentTrend, sentCol := trend(t.Sent, p.Sent, true)
	failTrend, failCol := trend(t.Failed, p.Failed, false)
	retryTrend, retryCol := trend(t.Retries, p.Retries, false)

	failedColor := ui.Color("")
	if t.Failed > 0 {
		failedColor = ui.SE0
	}

	waitingHint := "keine hängenden Mails"
	waitingCol := ui.SG0
	if stats.Queue.Stuck > 0 {
		waitingHint = fmt.Sprintf("%d älter als %s", stats.Queue.Stuck, formatDuration(stats.Options.StuckAfter))
		waitingCol = ui.SE0
	}

	rateDiff := (t.ErrorRate() - p.ErrorRate()) * 100
	rateHint := fmt.Sprintf("%+.1f Prozentpunkte", rateDiff)
	rateCol := ui.SG0
	if rateDiff > 0 {
		rateCol = ui.SE0
	}

	latencyHint := "vom Einreihen bis zum Versand"
	if p.AvgLatency() > 0 && t.AvgLatency() > 0 {
		latencyHint = "vorher " + formatDuration(p.AvgLatency())
	}

	firstHint := "vom Einreihen bis zum ersten Versuch"
	firstCol := ui.Color("")
	if p.AvgFirstAttemptLatency() > 0 && t.AvgFirstAttemptLatency() > 0 {
		firstHint = "vorher " + formatDuration(p.AvgFirstAttemptLatency())
		if t.AvgFirstAttemptLatency() > p.AvgFirstAttemptLatency() {
			firstCol = ui.SE0
		} else {
			firstCol = ui.SG0
		}
	}

	suppressedColor := ui.Color("")
	suppressedHint := "Spam-Schutz ohne Befund"
	if stats.Queue.Suppressed > 0 {
		suppressedColor = ui.SE0
		suppressedHint = "mögliche Endlosschleife"
	}

	return grid([5]int{2, 4, 4, 4, 8},
		kpi("Versendet", fmt.Sprint(t.Sent), "", sentTrend, sentCol),
		kpi("Wartend", fmt.Sprint(stats.Queue.Queued+stats.Queue.Error), "", waitingHint, waitingCol),
		kpi("Fehlversuche", fmt.Sprint(t.Failed), failedColor, failTrend, failCol),
		kpi("Wiederholungen", fmt.Sprint(t.Retries), "", retryTrend, retryCol),
		kpi("Fehlerquote", fmt.Sprintf("%.1f %%", t.ErrorRate()*100), "", rateHint, rateCol),
		kpi("Ø Zustellzeit", formatDuration(t.AvgLatency()), "", latencyHint, ""),
		kpi("Ø bis 1. Versuch", formatDuration(t.AvgFirstAttemptLatency()), "", firstHint, firstCol),
		kpi("Unterdrückt", fmt.Sprint(stats.Queue.Suppressed), suppressedColor, suppressedHint, suppressedColor),
	)
}

func throughputChart(wnd core.Window, stats mail.StatisticsResult, tr timeRange) core.View {
	loc := wnd.Location()
	start := stats.Options.From.In(loc)
	if tr.bucket == time.Hour {
		start = start.Truncate(time.Hour)
	} else {
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	}

	var slots []time.Time
	for t := start; t.Before(stats.Options.To); {
		slots = append(slots, t)
		if tr.bucket == time.Hour {
			t = t.Add(time.Hour)
		} else {
			t = t.AddDate(0, 0, 1)
		}
	}

	sent := make([]float64, len(slots))
	failed := make([]float64, len(slots))
	for _, b := range stats.Buckets {
		h := b.Hour.In(loc)
		for i := len(slots) - 1; i >= 0; i-- {
			if !h.Before(slots[i]) {
				sent[i] += float64(b.Sent)
				failed[i] += float64(b.Failed)
				break
			}
		}
	}

	labels := make([]string, len(slots))
	sentPoints := make([]chart.DataPoint, len(slots))
	failedPoints := make([]chart.DataPoint, len(slots))
	for i, s := range slots {
		if tr.bucket == time.Hour {
			labels[i] = s.Format("15:04")
		} else {
			labels[i] = s.Format("02.01.")
		}
		sentPoints[i] = chart.DataPoint{X: labels[i], Y: sent[i]}
		failedPoints[i] = chart.DataPoint{X: labels[i], Y: failed[i]}
	}

	return barchart.BarChart(chart.Chart{
		Labels:        labels,
		Colors:        []ui.Color{ui.SV0, ui.SE0},
		Frame:         ui.Frame{Height: ui.L320}.FullWidth(),
		NoDataMessage: "Keine Daten im gewählten Zeitraum",
		LabelRounding: chart.RoundingRound,
	}).Series([]chart.Series{
		{Label: "Versendet", Type: chart.ChartSeriesTypeColumn, DataPoints: sentPoints},
		{Label: "Fehlversuche", Type: chart.ChartSeriesTypeColumn, DataPoints: failedPoints},
	}).Stacked(true)
}

func statusDonut(stats mail.StatisticsResult) core.View {
	q := stats.Queue
	if q.Total() == 0 {
		return ui.Text("Die Warteschlange ist leer.").Color(ui.ST0)
	}

	return ui.VStack(
		piechart.PieChart(chart.Chart{
			Colors: []ui.Color{ui.SG0, ui.SV0, ui.SW0, ui.SE0, ui.ST0},
			Frame:  ui.Frame{Width: ui.L480, MaxWidth: ui.Full, Height: ui.L320},
		}).Series([]chart.Series{{
			Label: "Warteschlange",
			DataPoints: []chart.DataPoint{
				{X: "Versendet", Y: float64(q.Success)},
				{X: "Wartet", Y: float64(q.Queued)},
				{X: "Fehler, wird wiederholt", Y: float64(q.Error)},
				{X: "Endgültig fehlgeschlagen", Y: float64(q.Failed)},
				{X: "Unterdrückt", Y: float64(q.Suppressed)},
			},
		}}).ShowAsDonut(true).ShowAbsoluteValues(true),
		ui.Text(fmt.Sprintf("%d Mails in der Warteschlange", q.Total())).Font(ui.BodySmall).Color(ui.ST0),
	).FullWidth()
}

func recentFailures(wnd core.Window, pages Pages, stats mail.StatisticsResult) core.View {
	if len(stats.RecentFailures) == 0 {
		return ui.Text("Keine fehlgeschlagenen Mails.").Color(ui.SG0)
	}

	return ui.VStack(
		dataview.FromSlice(wnd, stats.RecentFailures, []dataview.Field[dataview.Element[mail.Outgoing]]{
			{Name: "Zeit", Map: func(e dataview.Element[mail.Outgoing]) core.View {
				return ui.Text(formatTime(wnd, e.Value.SendAt))
			}},
			{Name: "Empfänger", Map: func(e dataview.Element[mail.Outgoing]) core.View {
				return ui.Text(xstrings.EllipsisEnd(e.Value.Receiver, 28))
			}},
			{Name: "Server", Map: func(e dataview.Element[mail.Outgoing]) core.View {
				return ui.Text(e.Value.ServerName)
			}, Visible: dataview.MinSize2XL()},
			{Name: "Fehler", Map: func(e dataview.Element[mail.Outgoing]) core.View {
				return errorPill(e.Value)
			}},
		}).Action(func(e dataview.Element[mail.Outgoing]) {
			wnd.Navigation().ForwardTo(pages.OutgoingMail, core.Values{"id": string(e.Value.ID)})
		}).Selection(false).Search(false).Style(dataview.Table),
		ui.TertiaryButton(func() {
			wnd.Navigation().ForwardTo(pages.OutgoingMailQueue, core.Values{"status": string(mail.StatusError) + "," + string(mail.StatusFailed) + "," + string(mail.StatusSuppressed)})
		}).Title("Alle Fehler anzeigen"),
	).Alignment(ui.TopTrailing).Gap(ui.L8).FullWidth()
}

func errorPill(o mail.Outgoing) core.View {
	a, ok := o.LastAttempt()
	if !ok || a.Success {
		if o.LastError == "" {
			return ui.Text("–")
		}
		return tags.StatusBadge(ui.SE0, "Fehler")
	}

	txt := phaseLabel(a.Phase)
	if a.Code > 0 {
		txt = fmt.Sprintf("%d %s", a.Code, txt)
	}

	col := ui.SE0
	if o.Status == mail.StatusError {
		col = ui.SW0
	}

	return tags.StatusBadge(col, txt)
}
