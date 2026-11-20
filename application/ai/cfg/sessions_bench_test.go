// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
)

// The benchmarks of this file measure the conversations of nago's floating assistant in the stocks of the
// measurement M27 (wokoda, R-NFR-ASSISTANT-OPEN-TIME detail): the conversations of other persons and of the measuring
// person in nago's store, each of about 25 KB of text and tool calls, wired by [Enable] like in an application.
//
// Every measurement starts after closing and opening the data: a stock is seeded once into its directory and every
// run of a benchmark opens it with a fresh configurator. The directory is below $NAGO_BENCH_DATA, e.g. a volume of a
// container, otherwise in a temporary directory of the benchmark, which is seeded anew each time. Run them with
// -test.benchtime=50x for the median over 50 calls; the first call after opening is reported apart as first-ms.

// benchStock is a stock of conversations of M27.
type benchStock struct {
	name string
	// foreign is the number of conversations of other persons, benchPersonSessions per person
	foreign int
	// waiting tells, whether the first own conversation waits on a question
	waiting bool
}

var (
	stockS2  = benchStock{name: "s2", foreign: 2_000}
	stockS2W = benchStock{name: "s2w", foreign: 2_000, waiting: true}
	stockS50 = benchStock{name: "s50", foreign: 50_000}
	// stockS50W is S50 with a waiting own conversation
	stockS50W = benchStock{name: "s50w", foreign: 50_000, waiting: true}
)

const (
	// benchOwnSessions is the number of conversations of the measuring person
	benchOwnSessions = 200
	// benchPersonSessions is the number of conversations of each other person
	benchPersonSessions = 10
	benchAppID          = "de.worldiety.nago.bench.sessions"
	benchSeedFile       = "bench.json"
)

// benchSeed is what seeding a stock leaves for the measurements, in benchSeedFile of its directory.
type benchSeed struct {
	User         user.ID      `json:"user"`
	Secret       secret.ID    `json:"secret"`
	ProviderHint string       `json:"providerHint"`
	Sessions     int          `json:"sessions"`
	Own          []session.ID `json:"own"`
	SessionSize  int          `json:"sessionSize"`
}

// dir returns the directory of the stock.
func (s benchStock) dir(b *testing.B) string {
	root := os.Getenv("NAGO_BENCH_DATA")
	if root == "" {
		return filepath.Join(b.TempDir(), s.name)
	}

	return filepath.Join(root, s.name)
}

// seed writes the stock into dir, unless it is there already, and returns its seed. The conversations are written
// directly into the repository of nago.ai.session, like by a former process, and granted to their creators like
// [session.Create] does.
func (s benchStock) seed(b *testing.B, dir string) benchSeed {
	b.Helper()

	if buf, err := os.ReadFile(filepath.Join(dir, benchSeedFile)); err == nil {
		var seed benchSeed
		if err := json.Unmarshal(buf, &seed); err != nil {
			b.Fatal(err)
		}

		return seed
	}

	// a partial stock of an aborted seeding is written anew
	if err := os.RemoveAll(dir); err != nil {
		b.Fatal(err)
	}

	start := time.Now()
	app := nagotest.New(b, func(c *application.Configurator) {
		c.SetApplicationID(benchAppID)
		c.SetDataDir(dir)
	})
	cfg := app.Configurator()

	users := std.Must(cfg.UserManagement())
	usr := std.Must(users.UseCases.Create(user.SU(), user.ShortRegistrationUser{
		Firstname:        "Mira",
		Lastname:         "Messend",
		Email:            "mira.messend@example.com",
		Password:         "Sup3r-Geheim!2026",
		PasswordRepeated: "Sup3r-Geheim!2026",
		Verified:         true,
	}))

	// the provider of the conversations, which BenchmarkAssistantWindow points to its own server
	secrets := std.Must(cfg.SecretManagement())
	sid := std.Must(secrets.UseCases.CreateSecret(user.SU(), openai.Settings{Name: "Bench", BaseURL: "http://127.0.0.1:9/v1"}))
	if err := secrets.UseCases.UpdateMySecretGroups(user.SU(), sid, []group.ID{group.System}); err != nil {
		b.Fatal(err)
	}

	seed := benchSeed{User: usr.ID, Secret: sid, ProviderHint: string(sid)}

	// the wiring of Enable
	rdb := std.Must(cfg.RDB())
	rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Owner, Target: session.Namespace})
	for _, pid := range session.InstancePermissions {
		rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Relation(pid), Target: session.Namespace})
	}
	repo := session.Repository(std.Must(application.JSONRepository[session.Session](cfg, string(session.Namespace))))

	base := xtime.Now() - xtime.UnixMilliseconds(s.foreign+benchOwnSessions)*1000
	for i := range s.foreign {
		person := user.ID(benchIdent("person", i/benchPersonSessions))
		sess := benchSession(session.ID(benchIdent("foreign", i)), person, seed.ProviderHint, i, base)
		if i == 0 {
			benchWaiting(&sess)
		}
		if i == 1 {
			seed.SessionSize = len(std.Must(json.Marshal(sess)))
		}

		benchSave(b, repo, rdb, sess)
	}

	for i := range benchOwnSessions {
		sess := benchSession(session.ID(benchIdent("own", i)), usr.ID, seed.ProviderHint, s.foreign+i, base)
		if i == 0 && s.waiting {
			benchWaiting(&sess)
		}

		benchSave(b, repo, rdb, sess)
		seed.Own = append(seed.Own, sess.ID)
	}

	seed.Sessions = s.foreign + benchOwnSessions
	app.Core().Destroy()

	buf := std.Must(json.Marshal(seed))
	if err := os.WriteFile(filepath.Join(dir, benchSeedFile), buf, 0600); err != nil {
		b.Fatal(err)
	}

	b.Logf("seeded %s in %v: %d conversations, %d bytes of JSON each", s.name, time.Since(start), seed.Sessions, seed.SessionSize)
	return seed
}

// open opens the stock with a fresh configurator, which configure prepares, see built.
func (s benchStock) open(b *testing.B, extra int, configure func(c *application.Configurator)) (*nagotest.App, benchSeed) {
	b.Helper()

	seed := s.built(b, extra)
	dir := s.dir(b)
	app := benchOpenDir(b, dir, configure)
	repo := std.Must(application.JSONRepository[session.Session](app.Configurator(), string(session.Namespace)))
	if n := std.Must(repo.Count()); n != seed.Sessions+extra {
		b.Fatalf("stock %s has %d conversations instead of %d", s.name, n, seed.Sessions+extra)
	}

	return app, seed
}

// built seeds the stock and builds the heads of its conversations, with extra conversations besides the seeded ones,
// unless that has been done: a configurator with the ai management builds them, and every configurator is closed
// afterwards.
func (s benchStock) built(b *testing.B, extra int) benchSeed {
	b.Helper()

	dir := s.dir(b)
	seed := s.seed(b, dir)

	if !benchHeadsBuilt(b, dir, seed.Sessions+extra) {
		start := time.Now()
		app := benchOpenDir(b, dir, func(c *application.Configurator) { std.Must(Enable(c)) })
		benchAwaitHeads(b, app.Configurator(), seed.Sessions+extra)
		app.Core().Destroy()
		b.Logf("built the heads of %s in %v", s.name, time.Since(start))
	}

	return seed
}

// openEnabled opens the stock with the ai management enabled.
func (s benchStock) openEnabled(b *testing.B, extra int) (*nagotest.App, Management, benchSeed) {
	b.Helper()

	var mgmt Management
	app, seed := s.open(b, extra, func(c *application.Configurator) {
		mgmt = std.Must(Enable(c))
	})

	return app, mgmt, seed
}

// benchOpenDir opens the data directory with a fresh configurator.
func benchOpenDir(b *testing.B, dir string, configure func(c *application.Configurator)) *nagotest.App {
	b.Helper()

	return nagotest.New(b, func(c *application.Configurator) {
		c.SetApplicationID(benchAppID)
		c.SetDataDir(dir)
		configure(c)
	})
}

// The store of the heads and its keys, see application/ai/session/heads.go.
const (
	benchHeadStore = "nago.ai.session.head"
	benchLayoutKey = "~layout"
)

// benchHeads opens the store of the heads of the configurator.
func benchHeads(b *testing.B, cfg *application.Configurator) blob.Store {
	b.Helper()

	stores := std.Must(cfg.Stores())
	return std.Must(stores.Open(benchHeadStore, blob.OpenStoreOptions{Type: blob.EntityStore, Derived: true}))
}

// benchHeadsDone tells whether the build of the heads has finished, reading the layout only.
func benchHeadsDone(b *testing.B, heads blob.Store) bool {
	b.Helper()

	var l struct {
		State string `json:"state"`
	}
	if opt := std.Must(blob.Get(heads, benchLayoutKey)); opt.IsSome() {
		if err := json.Unmarshal(opt.Unwrap(), &l); err != nil {
			b.Fatal(err)
		}
	}

	return l.State == "done" && !std.Must(heads.Exists(context.Background(), blob.InvalidKey))
}

// benchHeadsState tells whether the build of the heads has finished and how many heads there are.
func benchHeadsState(b *testing.B, heads blob.Store) (done bool, n int) {
	b.Helper()

	for _, prefix := range []string{"h/", "k/"} {
		for _, err := range heads.List(context.Background(), blob.ListOptions{Prefix: prefix}) {
			if err != nil {
				b.Fatal(err)
			}
			n++
		}
	}

	return benchHeadsDone(b, heads), n
}

// benchHeadsBuilt opens the directory without the ai management and tells, whether the heads of all conversations
// are built.
func benchHeadsBuilt(b *testing.B, dir string, sessions int) bool {
	b.Helper()

	app := benchOpenDir(b, dir, func(c *application.Configurator) {})
	defer app.Core().Destroy()

	done, n := benchHeadsState(b, benchHeads(b, app.Configurator()))
	return done && n == sessions
}

// benchAwaitHeads waits until the build of the heads has finished with a head for every conversation.
func benchAwaitHeads(b *testing.B, cfg *application.Configurator, sessions int) {
	b.Helper()

	heads := benchHeads(b, cfg)
	for deadline := time.Now().Add(30 * time.Minute); !benchHeadsDone(b, heads); {
		if time.Now().After(deadline) {
			b.Fatalf("the build of the heads has not finished")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if _, n := benchHeadsState(b, heads); n != sessions {
		b.Fatalf("the build has finished with %d heads for %d conversations", n, sessions)
	}
}

// subject returns the measuring person of the stock.
func (s benchSeed) subject(b *testing.B, app *nagotest.App) auth.Subject {
	b.Helper()

	users := std.Must(app.Configurator().UserManagement())
	return std.Must(users.UseCases.SubjectFromUser(user.SU(), s.User)).Unwrap()
}

// benchIdent returns a fixed identifier of 32 hex digits like [data.RandIdent].
func benchIdent(kind string, i int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s/%d", kind, i)))
	return hex.EncodeToString(sum[:16])
}

// benchSession returns the i-th conversation of the stock, of about 25 KB.
func benchSession(id session.ID, createdBy user.ID, providerHint string, i int, base xtime.UnixMilliseconds) session.Session {
	return session.Session{
		ID:           id,
		Title:        fmt.Sprintf("Rechnungen prüfen %d", i),
		Model:        "bench-model",
		System:       "Du bist der Assistent einer Anwendung für Rechnungen.",
		ProviderHint: providerHint,
		Tags:         slices.Clone(benchTags),
		Messages:     benchMessages(i),
		CreatedBy:    createdBy,
		CreatedAt:    base + xtime.UnixMilliseconds(i)*1000,
		UpdatedAt:    base + xtime.UnixMilliseconds(i)*1000 + 500,
		Usage:        completion.Usage{InputTokens: 12_000, OutputTokens: 900},
	}
}

// benchWaiting lets the conversation wait on a question.
func benchWaiting(s *session.Session) {
	ask := completion.ToolCall{ID: "ask-0", Name: "ask_user", Arguments: json.RawMessage(`{"question":"Welche Rechnung soll ich prüfen?"}`)}
	s.Messages = append(s.Messages, completion.Message{Role: completion.Assistant, Content: []completion.Content{ask}})
	s.Pending = &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: ask, Question: "Welche Rechnung soll ich prüfen?"}}}
	s.PendingRevision = 1
}

// benchSave writes the conversation and grants it its creator like session.Create does.
func benchSave(b *testing.B, repo session.Repository, rdb *rebac.DB, s session.Session) {
	if err := repo.Save(s); err != nil {
		b.Fatal(err)
	}

	src := rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(s.CreatedBy)}
	target := rebac.Entity{Namespace: session.Namespace, Instance: rebac.Instance(s.ID)}
	relations := []rebac.Relation{rebac.Owner}
	for _, pid := range session.InstancePermissions {
		relations = append(relations, rebac.Relation(pid))
	}
	for _, rel := range relations {
		if err := rdb.Put(rebac.Triple{Source: src, Relation: rel, Target: target}); err != nil {
			b.Fatal(err)
		}
	}
}

// benchStocks are the stocks of G1, without and with a waiting own conversation.
var benchStocks = []benchStock{stockS2, stockS2W, stockS50, stockS50W}

// BenchmarkFindResumable is G1 of M27: the lookup of the conversation a freshly opened assistant continues, for the
// measuring person without and with a waiting conversation, in S2 and S50. Among the conversations of other persons,
// one waits on a question with the same tags and provider.
func BenchmarkFindResumable(b *testing.B) {
	for _, stock := range benchStocks {
		b.Run(stock.name, func(b *testing.B) {
			app, mgmt, seed := stock.openEnabled(b, 0)
			subject := seed.subject(b, app)

			var want session.ID
			if stock.waiting {
				want = seed.Own[0]
			}

			opts := session.FindResumableOptions{Tags: benchTags, ProviderHint: seed.ProviderHint}
			lookup := func() {
				found, err := mgmt.SessionUseCases.FindResumable(subject, opts)
				if err != nil {
					b.Fatal(err)
				}

				var got session.ID
				if found.IsSome() {
					got = found.Unwrap().ID
				}
				if got != want {
					b.Fatalf("found %q instead of %q", got, want)
				}
			}

			benchMeasure(b, lookup)
		})
	}
}

// BenchmarkHistory is G2 of M27: loading the heads of the history of the measuring person, the 200 own conversations,
// in S2 and S50, through FindSummaries like the history dialog.
func BenchmarkHistory(b *testing.B) {
	for _, stock := range []benchStock{stockS2, stockS50} {
		b.Run(stock.name, func(b *testing.B) {
			app, mgmt, seed := stock.openEnabled(b, 0)
			subject := seed.subject(b, app)

			load := func() {
				n := 0
				for s, err := range mgmt.SessionUseCases.FindSummaries(subject, session.FindAllOptions{Tags: benchTags, CreatedBy: seed.User}) {
					if err != nil {
						b.Fatal(err)
					}
					if s.CreatedBy != seed.User {
						b.Fatalf("history has a conversation of %s", s.CreatedBy)
					}
					n++
				}

				if n != benchOwnSessions {
					b.Fatalf("history has %d conversations instead of %d", n, benchOwnSessions)
				}
			}

			benchMeasure(b, load)
		})
	}
}

// BenchmarkDelete is G7 of M27: deleting a conversation without children in S50. Before opening, b.N+1 conversations
// of about 25 KB are created through the use cases, which the measurement deletes again, so that the stock stays.
func BenchmarkDelete(b *testing.B) {
	stock := stockS50

	// created with a configurator of its own, so that the measurement starts after opening
	app, mgmt, seed := stock.openEnabled(b, 0)
	text := strings.Repeat("Die Rechnung 4711 enthält drei Positionen, einen Rabatt und die Umsatzsteuer. ", 330)
	var ids []session.ID
	for range b.N + 1 {
		s, err := mgmt.SessionUseCases.Create(user.SU(), session.CreateOptions{
			Title:        "Zu löschen",
			Tags:         benchTags,
			ProviderHint: seed.ProviderHint,
			Input:        []completion.Content{completion.Text{Text: text}},
		})
		if err != nil {
			b.Fatal(err)
		}
		ids = append(ids, s.ID)
	}
	app.Core().Destroy()

	_, mgmt, _ = stock.openEnabled(b, len(ids))

	next := 0
	del := func() {
		id := ids[next]
		next++
		if err := mgmt.SessionUseCases.Delete(user.SU(), id); err != nil {
			b.Fatal(err)
		}
		if found := std.Must(mgmt.SessionUseCases.FindByID(user.SU(), id)); found.IsSome() {
			b.Fatalf("%s is not deleted", id)
		}
	}

	benchMeasure(b, del)
}

// benchMeasure times the first call of fn apart, then b.N calls one by one, and reports first-ms and median-ms.
func benchMeasure(b *testing.B, fn func()) {
	b.Helper()

	first := benchTimed(fn)

	durations := make([]time.Duration, 0, b.N)
	b.ResetTimer()
	for range b.N {
		durations = append(durations, benchTimed(fn))
	}
	b.StopTimer()

	if len(durations) != b.N {
		b.Fatalf("measured %d of %d calls", len(durations), b.N)
	}

	b.ReportMetric(benchMillis(first), "first-ms")
	b.ReportMetric(benchMillis(benchMedian(durations)), "median-ms")
}

func benchTimed(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

func benchMillis(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
