// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Tutorial 103 – ndb: Message-Streams (Event Sourcing) und Zeitreihen (tsdb).
//
// Dieses Beispiel zeigt die beiden Kern-Anwendungsfälle der ndb-Datenbank an je
// einer Seite:
//
//  1. "/"        Nachrichten-/Event-Sourcing-Demo über ndb.Messages (msgstore).
//     Sie zeigt die zwei gebräuchlichen Lesemuster auf demselben
//     Event-Log und erklärt, wann man welches nimmt:
//     - evs.Handler  (decide/evolve): der Schreibpfad. Er lädt jede
//     Aggregat-Instanz in den Speicher, wendet Kommandos über
//     Decide an, persistiert die entstandenen Events und faltet
//     sie über Evolve wieder ein. Nutze ihn, wenn du *entscheiden*
//     musst (Invarianten, Validierung) bevor ein Event entsteht.
//     - evs.Projection: der Lesepfad / das Read-Model. Sie liest
//     denselben Event-Strom (History + Notifier) und faltet ihn zu
//     beliebig vielen, unabhängigen Sichten. Nutze sie für
//     Auswertungen, Listen, Zähler, Dashboards – überall, wo du
//     *nur lesen* willst und mehrere Sichten pro Event brauchst.
//
//  2. "/tsdb"    Zeitreihen-Demo über die tsdb-Engine. Sie schreibt eine
//     oszillierende 50-Hz-Reihe, liest sie als iter.Seq[Point] und
//     reduziert sie mit timeseries.M4 auf die Pixelbreite der
//     Anzeige, bevor sie als LineChart gerendert wird. So bleibt die
//     Darstellung auch bei Milliarden Punkten konstant im Aufwand.
package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/evs"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/cloner"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
	"go.wdy.de/nago/pkg/ndb/tsdb"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/timeseries"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/chart"
	"go.wdy.de/nago/presentation/ui/linechart"
	"go.wdy.de/nago/web/vuejs"
)

// ============================================================================
// Domäne für die Message-/Event-Sourcing-Demo
// ============================================================================

// AccountID ist der Aggregat-Schlüssel: ein Bankkonto.
type AccountID string

// Account ist das Aggregat (der In-Memory-Zustand). evs verlangt Clone() für
// race-freie Snapshots und IsDeleted() für logische Löschung.
type Account struct {
	ID      AccountID
	Balance int64 // in Cent
	Deleted bool
}

func (a *Account) Clone() *Account {
	if a == nil {
		return nil
	}
	c := *a
	return &c
}

func (a *Account) IsDeleted() bool { return a.Deleted }

// AccEvt ist der Summentyp aller Konto-Events.
type AccEvt = evs.Evt[*Account]

// MoneyDeposited / MoneyWithdrawn sind die Fakten. Jedes Event trägt seinen
// stabilen Discriminator (= ndb.TypeID/Stream-Name) und seine Evolve-Funktion,
// die den Aggregat-Zustand fortschreibt.
type MoneyDeposited struct {
	Account AccountID `json:"account"`
	Amount  int64     `json:"amount"`
}

func (MoneyDeposited) Discriminator() evs.Discriminator { return "MoneyDeposited" }

func (e MoneyDeposited) Evolve(_ context.Context, a *Account) error {
	a.ID = e.Account
	a.Balance += e.Amount
	return nil
}

type MoneyWithdrawn struct {
	Account AccountID `json:"account"`
	Amount  int64     `json:"amount"`
}

func (MoneyWithdrawn) Discriminator() evs.Discriminator { return "MoneyWithdrawn" }

func (e MoneyWithdrawn) Evolve(_ context.Context, a *Account) error {
	a.ID = e.Account
	a.Balance -= e.Amount
	return nil
}

// DepositCmd / WithdrawCmd sind Kommandos. Ihre Decide-Methode ist der Ort für
// Invarianten: hier wird *entschieden*, ob (und welche) Events entstehen dürfen.
// Genau dafür ist der Handler da – nicht für reines Lesen.
type DepositCmd struct {
	ID     AccountID
	Amount int64
}

func (c DepositCmd) Decide(_ auth.Subject, _ *Account) ([]AccEvt, error) {
	if c.Amount <= 0 {
		return nil, fmt.Errorf("Betrag muss positiv sein")
	}
	return []AccEvt{MoneyDeposited{Account: c.ID, Amount: c.Amount}}, nil
}

type WithdrawCmd struct {
	ID     AccountID
	Amount int64
}

func (c WithdrawCmd) Decide(_ auth.Subject, a *Account) ([]AccEvt, error) {
	if c.Amount <= 0 {
		return nil, fmt.Errorf("Betrag muss positiv sein")
	}
	if a.Balance < c.Amount {
		// Invariante: kein Dispo. Diese Prüfung braucht den aktuellen Zustand –
		// deshalb der Handler und nicht eine Projection.
		return nil, fmt.Errorf("nicht genügend Guthaben (%d < %d)", a.Balance, c.Amount)
	}
	return []AccEvt{MoneyWithdrawn{Account: c.ID, Amount: c.Amount}}, nil
}

// accountID routet ein Event auf seinen Aggregat-Schlüssel.
func accountID(e AccEvt) (AccountID, bool) {
	switch evt := e.(type) {
	case MoneyDeposited:
		return evt.Account, evt.Account != ""
	case MoneyWithdrawn:
		return evt.Account, evt.Account != ""
	default:
		return "", false
	}
}

// ---- Read-Model (Projection) --------------------------------------------------

// ledgerStats ist eine globale Auswertung über *alle* Konten, gefaltet aus
// demselben Event-Strom. Genau dafür ist eine Projection da: eine unabhängige
// Sicht, die man beliebig oft und in beliebiger Form neben dem Handler betreiben
// kann, ohne den Schreibpfad zu berühren.
type ledgerStats struct {
	Deposits    int64 // Summe aller Einzahlungen (Cent)
	Withdrawals int64 // Summe aller Auszahlungen (Cent)
	Count       int64 // Anzahl gefalteter Events
}

func (s *ledgerStats) Clone() *ledgerStats {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}

var _ cloner.Cloner[*ledgerStats] = (*ledgerStats)(nil)

// newLedgerProjection baut eine Singleton-Projection (ein Schlüssel) über die
// beiden Event-Typen. src ist der ndb-Message-Store selbst (History+Notifier).
func newLedgerProjection(src evs.Source) *evs.Singleton[*ledgerStats] {
	p := evs.NewProjection[evs.Unit, *ledgerStats](src, evs.ProjectionOptions{})
	evs.Project(p,
		func(MoneyDeposited) evs.Unit { return evs.TheUnit() },
		func(s *ledgerStats, e MoneyDeposited) { s.Deposits += e.Amount; s.Count++ },
	)
	evs.Project(p,
		func(MoneyWithdrawn) evs.Unit { return evs.TheUnit() },
		func(s *ledgerStats, e MoneyWithdrawn) { s.Withdrawals += e.Amount; s.Count++ },
	)
	return p
}

// ============================================================================
// Verdrahtung
// ============================================================================

const (
	pageMessages core.NavigationPath = "."
	pageTSDB     core.NavigationPath = "tsdb"

	tsBucket    = "demo"
	tsColumn    = "sensor"
	tsColumnStr = "status" // string-Zeitreihe zum Testen des String/Enum-Falls
	tsStepMs    = 20       // 50 Hz
	tsCount     = 200_000
	tsStrStepMs = 1000 // 1 Hz Statuswechsel
	tsStrCount  = 4_000
)

// app hält die zur Laufzeit gemeinsam genutzten Objekte.
type app struct {
	handler     *evs.Handler[*Account, AccEvt, AccountID]
	ledger      *evs.Singleton[*ledgerStats]
	tsColumn    *tsdb.Column
	tsColumnStr *tsdb.Column
	tsBaseMs    int64
	tsLastMs    int64
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_103")
		cfg.Serve(vuejs.Dist())

		option.Must(cfginspector.Enable(cfg))
		// Admin-Auth + Standard-Systeme wie in den anderen Beispielen.
		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		a := option.Must(setup(cfg))

		cfg.RootViewWithDecoration(pageMessages, func(wnd core.Window) core.View {
			return pageMessagesView(wnd, a)
		})
		cfg.RootViewWithDecoration(pageTSDB, func(wnd core.Window) core.View {
			return pageTSDBView(wnd, a)
		})
	}).Run()
}

// setup öffnet die gemeinsame ndb-Datenbank über die Configurator-Fabrik,
// erzeugt darin eine msgstore- und eine tsdb-Engine und seedet Demodaten.
func setup(cfg *application.Configurator) (*app, error) {
	db, err := cfg.NDB() // gemeinsame, automatisch geschlossene ndb-Instanz unter DataDir()/ndb
	if err != nil {
		return nil, err
	}

	// ---- Message-Store (msgstore) einrichten -------------------------------
	msgEng, err := db.Engine("accounts", ndb.EngineOptions{
		Kind:   msgstore.EngineKind,
		Config: msgstore.Options{},
	})
	if err != nil {
		return nil, err
	}
	msgs := msgEng.(ndb.MessageEngine).Messages()

	// Der Handler (Schreibpfad) sitzt über einem ndb-Backend auf demselben
	// Message-Store. NewNDBBackend erzeugt aus ndb.Messages ein evs.Backend.
	backend := evs.NewNDBBackend[AccEvt, *Account](msgs)
	handler := evs.NewHandler[*Account](backend, accountID, backend.Register)
	handler.RegisterEvents(MoneyDeposited{}, MoneyWithdrawn{})

	// Die Projection (Lesepfad) liest denselben Strom direkt als Source
	// (ndb.Messages ist ndb.Followable = History + Notifier).
	ledger := newLedgerProjection(msgs)
	ledger.Run() // startet Warmup + Live-Tail im Hintergrund

	// ---- Zeitreihe (tsdb) einrichten ---------------------------------------
	tsEng, err := db.Engine("metrics", ndb.EngineOptions{
		Kind:   tsdb.EngineKind,
		Config: tsdb.Options{},
	})
	if err != nil {
		return nil, err
	}
	tdb := tsEng.(interface{ DB() *tsdb.DB }).DB()
	col, err := tdb.Column(tsBucket, tsColumn, tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 2})
	if err != nil {
		return nil, err
	}

	// zweite Zeitreihe: eine String-Spalte für den String/Enum-Fall im Inspektor.
	colStr, err := tdb.Column(tsBucket, tsColumnStr, tsdb.Schema{Scheme: tsdb.SchemeString})
	if err != nil {
		return nil, err
	}

	a := &app{handler: handler, ledger: ledger, tsColumn: col, tsColumnStr: colStr}

	// Zeitreihen nur einmalig seeden (falls die Spalten noch leer sind).
	a.tsBaseMs = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	a.tsLastMs = a.tsBaseMs + int64(tsCount-1)*tsStepMs
	if isColumnEmpty(col, a.tsBaseMs, a.tsLastMs) {
		if err := seedTimeseries(col, a.tsBaseMs); err != nil {
			return nil, err
		}
	}
	strLast := a.tsBaseMs + int64(tsStrCount-1)*tsStrStepMs
	if isColumnEmptyStr(colStr, a.tsBaseMs, strLast) {
		if err := seedStatusTimeseries(colStr, a.tsBaseMs); err != nil {
			return nil, err
		}
	}

	return a, nil
}

func isColumnEmpty(col *tsdb.Column, min, max int64) bool {
	empty := true
	_ = col.ScanF64(min, max, func(ts []int64, _ []float64) bool {
		if len(ts) > 0 {
			empty = false
		}
		return false // erste Charge genügt
	})
	return empty
}

// seedTimeseries schreibt eine oszillierende 50-Hz-Reihe (Sinus + leichtes
// Rauschen). PutF64 skaliert transparent auf die konfigurierten Nachkommastellen.
func seedTimeseries(col *tsdb.Column, baseMs int64) error {
	for i := 0; i < tsCount; i++ {
		ts := baseMs + int64(i)*tsStepMs
		v := 20.0 + 5.0*math.Sin(float64(i)/200.0) + 0.5*math.Sin(float64(i)/7.0)
		if err := col.PutF64(ts, v); err != nil {
			return err
		}
	}
	return col.Flush()
}

func isColumnEmptyStr(col *tsdb.Column, min, max int64) bool {
	empty := true
	_ = col.ScanString(min, max, func(ts []int64, _ []string) bool {
		if len(ts) > 0 {
			empty = false
		}
		return false
	})
	return empty
}

// seedStatusTimeseries schreibt eine String-Zeitreihe mit wechselnden
// Statuswerten (1 Hz) für den String/Enum-Fall des Zeitreihen-Inspektors.
func seedStatusTimeseries(col *tsdb.Column, baseMs int64) error {
	states := []string{"idle", "running", "warning", "error", "maintenance"}
	for i := 0; i < tsStrCount; i++ {
		ts := baseMs + int64(i)*tsStrStepMs
		// die meiste Zeit "running", gelegentlich andere Zustände
		s := states[1]
		switch {
		case i%137 == 0:
			s = states[3] // error
		case i%53 == 0:
			s = states[2] // warning
		case i%311 == 0:
			s = states[4] // maintenance
		case i%17 == 0:
			s = states[0] // idle
		}
		if err := col.PutString(ts, s); err != nil {
			return err
		}
	}
	return col.Flush()
}

// ============================================================================
// Seite 1: Nachrichten / Event Sourcing
// ============================================================================

func pageMessagesView(wnd core.Window, a *app) core.View {
	accID := core.AutoState[string](wnd).Init(func() string { return "acc-1" })
	amount := core.AutoState[int64](wnd).Init(func() int64 { return 1000 })
	invalidate := core.AutoState[int](wnd)

	// aktuellen Aggregat-Zustand über den Handler lesen (Schreibmodell-Sicht)
	acc, _ := a.handler.Aggregate(wnd.Context(), AccountID(accID.Get()))
	balance := int64(0)
	if acc != nil {
		balance = acc.Balance
	}

	// Read-Model-Sicht über die Projection: die Projection ist eventual
	// consistent – wir warten kurz auf den Tail, damit die Demo deterministisch
	// wirkt. In echt liest man einfach den zuletzt bekannten Wert.
	stats, _ := evs.Value(a.ledger)
	if stats == nil {
		stats = &ledgerStats{}
	}

	run := func(cmd evs.Cmd[*Account, AccEvt]) {
		seq, err := a.handler.Handle(user.SU(), AccountID(accID.Get()), cmd)
		if err != nil {
			ui.Text(err.Error()) // in echt: alert.ShowBannerError(wnd, err)
			return
		}
		// Auf das Read-Model warten, damit die Zahlen unten sofort stimmen.
		_ = a.ledger.WaitFor(wnd.Context(), ndb.Seq(seq))
		invalidate.Set(invalidate.Get() + 1)
	}

	return ui.VStack(
		ui.H1("ndb · Event Sourcing"),
		ui.Text("Konto (Aggregat-Schlüssel), zwei Kommandos, ein Event-Log – gelesen auf zwei Arten.").
			Font(ui.BodyLarge),

		ui.HStack(
			ui.SecondaryButton(func() { wnd.Navigation().ForwardTo(pageTSDB, nil) }).
				Title("→ Zeitreihen-Demo"),
		).FullWidth().Alignment(ui.Trailing),

		ui.Space(ui.L16),

		// Eingaben
		ui.HStack(
			ui.TextField("Konto", accID.Get()).InputValue(accID).Frame(ui.Frame{Width: ui.L160}),
			ui.IntField("Betrag (Cent)", amount.Get(), amount).Frame(ui.Frame{Width: ui.L160}),
			ui.PrimaryButton(func() { run(DepositCmd{ID: AccountID(accID.Get()), Amount: amount.Get()}) }).
				Title("Einzahlen"),
			ui.SecondaryButton(func() { run(WithdrawCmd{ID: AccountID(accID.Get()), Amount: amount.Get()}) }).
				Title("Auszahlen"),
		).Gap(ui.L8).FullWidth().Alignment(ui.Bottom),

		ui.Space(ui.L16),

		// Schreibmodell-Sicht (Handler)
		explainerCard(
			"Handler (decide/evolve) – der Schreibpfad",
			"Der Handler lädt das Aggregat in den Speicher, prüft in Decide die Invarianten "+
				"(z. B. „kein Dispo“) und persistiert nur dann Events. Nutze ihn immer dann, "+
				"wenn eine Entscheidung den aktuellen Zustand braucht, bevor ein Event entstehen darf.",
			ui.Text(fmt.Sprintf("Kontostand %q: %s", accID.Get(), euro(balance))).Font(ui.Title),
		),

		ui.Space(ui.L8),

		// Lesemodell-Sicht (Projection)
		explainerCard(
			"Projection – der Lesepfad / das Read-Model",
			"Die Projection faltet denselben Event-Strom unabhängig zu einer beliebigen Sicht. "+
				"Nutze sie für Auswertungen, Listen, Zähler oder Dashboards – überall, wo du nur "+
				"liest und ggf. mehrere Sichten pro Event brauchst. Sie berührt den Schreibpfad nicht.",
			ui.VStack(
				kvRow("Einzahlungen gesamt", euro(stats.Deposits)),
				kvRow("Auszahlungen gesamt", euro(stats.Withdrawals)),
				kvRow("Gefaltete Events", fmt.Sprintf("%d", stats.Count)),
			).Alignment(ui.Leading).FullWidth(),
		),
	).Gap(ui.L8).Alignment(ui.Leading).FullWidth()
}

// ============================================================================
// Seite 2: Zeitreihen + M4 + LineChart
// ============================================================================

func pageTSDBView(wnd core.Window, a *app) core.View {
	// Pixelbreite der Zeichenfläche grob aus dem Fenster ableiten. M4 liefert je
	// Bucket bis zu 4 Punkte (min/max/first/last), also ~4·width Punkte gesamt –
	// unabhängig davon, ob die Reihe Tausende oder Milliarden Rohpunkte hat.
	width := int(wnd.Info().Width) / 4
	if width < 50 {
		width = 50
	}
	if width > 400 {
		width = 400
	}

	// Fenster wählbar: gesamte Reihe oder ein enger Ausschnitt, um M4 zu zeigen.
	full := core.AutoState[bool](wnd).Init(func() bool { return true })
	min, max := a.tsBaseMs, a.tsLastMs
	if !full.Get() {
		// erste 5 Sekunden (250 Punkte) – hier sieht man Rohdaten fast 1:1
		max = min + 5000
	}

	rng := timeseries.NewRange(timeseries.UnixMilli(min), timeseries.UnixMilli(max), time.UTC)

	// M4 konsumiert den tsdb-Iterator lazy und reduziert konstant im Speicher.
	var pts []chart.DataPoint
	for p := range timeseries.M4(a.tsColumn.IterF64(min, max), rng, width) {
		pts = append(pts, chart.DataPoint{
			X: time.UnixMilli(int64(p.X)).UTC().Format("15:04:05.000"),
			Y: float64(p.Y),
		})
	}

	series := []chart.Series{{
		Label:      "sensor",
		Type:       chart.ChartSeriesTypeLine,
		DataPoints: pts,
	}}

	c := chart.Chart{
		Frame:      ui.Frame{Height: ui.L400}.FullWidth(),
		XAxisTitle: "Zeit",
		YAxisTitle: "Wert",
	}

	return ui.VStack(
		ui.H1("ndb · Zeitreihen (tsdb) + M4"),
		ui.Text(fmt.Sprintf(
			"%d Rohpunkte (50 Hz). M4 reduziert auf ~%d Buckets (%d gezeichnete Punkte).",
			tsCount, width, len(pts))).Font(ui.BodyLarge),

		ui.HStack(
			ui.SecondaryButton(func() { wnd.Navigation().ForwardTo(pageMessages, nil) }).
				Title("← Nachrichten-Demo"),
			ui.Spacer(),
			ui.SecondaryButton(func() { full.Set(!full.Get()) }).
				Title(map[bool]string{true: "Ausschnitt zeigen", false: "Gesamte Reihe"}[full.Get()]),
		).FullWidth(),

		ui.Space(ui.L16),

		explainerCard(
			"M4 – visualisierungsorientiertes Downsampling",
			"M4 teilt den Zeitbereich in „width“ Buckets und liefert je "+
				"Bucket nur die vier für ein Liniendiagramm sichtbaren Punkte: erster, letzter, "+
				"Minimum und Maximum. Der Aufwand bleibt konstant zur Pixelbreite – so lassen sich "+
				"auch Milliarden Punkte flüssig darstellen. tsdb liefert die Rohdaten als "+
				"iter.Seq[Point], den M4 direkt und speicherschonend konsumiert.",
			linechart.LineChart(c).Curve(linechart.CurveSmooth).Series(series),
		),

		ui.Space(ui.L8),

		explainerCard(
			"String-Zeitreihe (Status)",
			"tsdb speichert neben Dezimalwerten auch String-/Enum-Spalten. Diese lassen sich "+
				"nicht als Chart darstellen, sondern werden im Zeitreihen-Inspektor als "+
				"fensterweise Werteliste angezeigt. Unten die ersten Statuswechsel.",
			statusTable(a),
		),
	).Gap(ui.L8).Alignment(ui.Leading).FullWidth()
}

// statusTable zeigt die ersten Werte der String-Zeitreihe als kleine Tabelle.
func statusTable(a *app) core.View {
	strLast := a.tsBaseMs + int64(tsStrCount-1)*tsStrStepMs
	type row struct {
		ts int64
		v  string
	}
	var rows []row
	_ = a.tsColumnStr.ScanString(a.tsBaseMs, strLast, func(ts []int64, vals []string) bool {
		for i := range ts {
			rows = append(rows, row{ts[i], vals[i]})
			if len(rows) >= 20 {
				return false
			}
		}
		return len(rows) < 20
	})

	trows := make([]ui.TTableRow, 0, len(rows))
	for _, r := range rows {
		trows = append(trows, ui.TableRow(
			ui.TableCell(ui.Text(time.UnixMilli(r.ts).UTC().Format("15:04:05"))),
			ui.TableCell(ui.Text(r.v)),
		))
	}
	return ui.Table(
		ui.TableColumn(ui.Text("Zeit")),
		ui.TableColumn(ui.Text("Status")),
	).Rows(trows...).Frame(ui.Frame{}.FullWidth())
}

// ============================================================================
// kleine UI-Helfer
// ============================================================================

func explainerCard(title, text string, body core.View) core.View {
	return ui.VStack(
		ui.Text(title).Font(ui.Title),
		ui.Text(text),
		ui.Space(ui.L8),
		body,
	).Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardBody).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(ui.Frame{}.FullWidth())
}

func kvRow(k, v string) core.View {
	return ui.HStack(
		ui.Text(k+":").Frame(ui.Frame{Width: ui.L256}),
		ui.Text(v).Font(ui.BodyLarge),
	).Alignment(ui.Leading).FullWidth()
}

func euro(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d,%02d €", sign, cents/100, cents%100)
}
