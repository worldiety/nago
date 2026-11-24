// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/mail/nms"
	"go.wdy.de/nago/application/mail/nms/nmstest"
	"go.wdy.de/nago/application/onboarding"
	cfgonboarding "go.wdy.de/nago/application/onboarding/cfg"
	uionboarding "go.wdy.de/nago/application/onboarding/ui"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

const address = "torben@example.com"

// routes of the test application besides the index
var pageRoutes = []core.NavigationPath{"impressum", "datenschutz", "page/agb", "extra", "other"}

// pageApp installs the setup with the options. Prepare runs before, e.g. to store the theme settings.
func pageApp(t *testing.T, opts cfgonboarding.Options, prepare func(cfg *application.Configurator)) (*nagotest.App, cfgonboarding.Management) {
	t.Helper()
	var management cfgonboarding.Management
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.onboardingtest")
		cfg.SetName("Testapp")
		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.Text("home")
		})

		for _, route := range pageRoutes {
			cfg.RootView(route, func(wnd core.Window) core.View {
				return ui.Text("page " + string(route))
			})
		}

		if prepare != nil {
			prepare(cfg)
		}

		m, err := cfgonboarding.Enable(cfg, opts)
		if err != nil {
			t.Fatal(err)
		}

		management = m
	})

	return app, management
}

func (m *mailbox) count() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.codes)
}

// switchable is a delivery, which a test changes.
type switchable struct {
	mutex    sync.Mutex
	delivery onboarding.Delivery
}

func level(l onboarding.DeliveryLevel) *switchable {
	return &switchable{delivery: onboarding.Delivery{Level: l}}
}

func (s *switchable) check() onboarding.Delivery {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.delivery
}

func (s *switchable) set(d onboarding.Delivery) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.delivery = d
}

// shareSMTP stores an SMTP secret, which is shared with the given groups.
func shareSMTP(t *testing.T, cfg *application.Configurator, smtp secret.SMTP, groups ...group.ID) {
	t.Helper()
	secrets := std.Must(cfg.SecretManagement())
	if _, err := secrets.UseCases.CreateSecret(user.SU(), smtp, secret.CreateOptions{Groups: groups}); err != nil {
		t.Fatal(err)
	}
}

// allText returns every text and input value of the window.
func allText(w *nagotest.Window) string {
	var sb strings.Builder
	for _, n := range w.FindAll(nagotest.Type[*proto.TextView]()).Nodes() {
		sb.WriteString(string(n.Component.(*proto.TextView).Value) + "\n")
	}

	for _, n := range w.FindAll(nagotest.Type[*proto.TextField]()).Nodes() {
		sb.WriteString(string(n.Component.(*proto.TextField).Value) + "\n")
	}

	return sb.String()
}

func clock(t time.Time) string {
	return t.In(time.Local).Format("15:04")
}

// assertProblemPage checks the parts, which every problem page has, and no way to request a code.
func assertProblemPage(t *testing.T, w *nagotest.Window) {
	t.Helper()
	assertProblemPageOf(t, w, onboarding.EnvInitialUserEmail)
}

// assertProblemPageOf is assertProblemPage with the variable of the address.
func assertProblemPageOf(t *testing.T, w *nagotest.Window, emailEnv string) {
	t.Helper()
	w.Find(nagotest.Text("Einrichtung noch nicht möglich")).Exactly(1)
	w.FindAll(nagotest.Type[*proto.Accordion]()).Exactly(1)
	ops := w.Find(nagotest.ID(uionboarding.IDOperators))
	ops.Find(nagotest.Text("Für den Betrieb")).Exactly(1)
	ops.FindAll(nagotest.TextContains("fehlt oder ist ungültig.")).Exactly(1)
	ops.Find(nagotest.Text("• " + emailEnv + " fehlt oder ist ungültig.")).Exactly(1)
	ops.Find(nagotest.Text("• Nago Mail Service nicht verfügbar: öffentliche https-Adresse nötig, Token fehlt oder " + application.EnvMailService + "=off.")).Exactly(1)
	ops.Find(nagotest.Text("• Kein SMTP-Konto in der Systemgruppe; es lässt sich erst nach dem ersten Konto hinterlegen.")).Exactly(1)
	w.Find(nagotest.ID(uionboarding.IDRecheck)).Find(nagotest.Text("Erneut prüfen")).Exactly(1)

	// no code field and no button, which requests a code
	w.FindAll(nagotest.ID("onboarding-code")).None()
	w.FindAll(nagotest.Type[*proto.TextField]()).None()
	w.FindAll(nagotest.Text("Einrichtung starten")).None()
	w.FindAll(nagotest.Text("Code erneut senden")).None()
}

// The problem page names the reasons in a fixed order and the times of a failed token exchange, only without a
// delivery. With a delivery and a valid address, the welcome follows instead.
func TestProblemPage(t *testing.T) {
	failed := time.Date(2026, 11, 24, 10, 5, 0, 0, time.Local)
	exchange := onboarding.Delivery{Level: onboarding.DeliveryNone, ExchangeFailedAt: failed, NextExchangeAt: failed.Add(time.Hour)}
	noDelivery := "Diese Anwendung kann noch keine E-Mails versenden. Der Code für das erste Konto kommt per E-Mail."
	noAddress := "Für diese Anwendung ist keine gültige Adresse für das erste Konto eingetragen."
	contact := "Bitte wende dich an den Betrieb dieser Anwendung."
	for _, tc := range []struct {
		name     string
		email    string
		delivery onboarding.Delivery
		banner   string // empty is no problem
		exchange string
	}{
		{name: "no delivery", email: address, delivery: onboarding.Delivery{Level: onboarding.DeliveryNone}, banner: noDelivery + " " + contact},
		{name: "no address", email: "no-mail", delivery: onboarding.Delivery{Level: onboarding.DeliveryAssured}, banner: noAddress + " " + contact},
		{name: "no address and no delivery", email: "  ", delivery: onboarding.Delivery{Level: onboarding.DeliveryNone}, banner: noDelivery + " " + noAddress + " " + contact},
		{name: "failed exchange", email: address, delivery: exchange, banner: noDelivery + " " + contact,
			exchange: "Letzter Anmeldeversuch beim Nago Mail Service gescheitert: " + clock(failed) + " Uhr. Nächster Versuch frühestens " + clock(failed.Add(time.Hour)) + " Uhr."},
		{name: "failed exchange, but tryable", email: "no-mail", delivery: onboarding.Delivery{Level: onboarding.DeliveryTryable, ExchangeFailedAt: failed, NextExchangeAt: failed.Add(time.Hour)}, banner: noAddress + " " + contact},
		{name: "assured", email: address, delivery: onboarding.Delivery{Level: onboarding.DeliveryAssured}},
		{name: "tryable", email: address, delivery: onboarding.Delivery{Level: onboarding.DeliveryTryable, ExchangeFailedAt: failed, NextExchangeAt: failed.Add(time.Hour)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			delivery := &switchable{delivery: tc.delivery}
			app, _ := pageApp(t, cfgonboarding.Options{Email: tc.email, Deliver: (&mailbox{}).deliver, Delivery: delivery.check}, nil)
			w := app.Open(t, nil, ".")

			if tc.banner == "" {
				w.FindAll(nagotest.ID(uionboarding.IDProblem)).None()
				w.FindAll(nagotest.ID(uionboarding.IDRecheck)).None()
				w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
				return
			}

			assertProblemPage(t, w)
			if got := w.Find(nagotest.ID(uionboarding.IDProblem)).Text(); got != tc.banner {
				t.Fatalf("expected the banner %q, got %q", tc.banner, got)
			}

			if tc.exchange == "" {
				w.FindAll(nagotest.ID(uionboarding.IDExchangeFailed)).None()
				return
			}

			if got := w.Find(nagotest.ID(uionboarding.IDExchangeFailed)).Text(); got != tc.exchange {
				t.Fatalf("expected %q, got %q", tc.exchange, got)
			}
		})
	}
}

// Without a mail service, only an SMTP server, which is not empty and shared with the system group, is a delivery.
// The recheck reads the secrets anew.
func TestRecheckReadsTheSecretsAnew(t *testing.T) {
	t.Setenv(application.EnvMailService, "off")
	app, _ := pageApp(t, cfgonboarding.Options{Email: address}, nil)
	w := app.Open(t, nil, ".")
	assertProblemPage(t, w)

	for _, tc := range []struct {
		name   string
		smtp   secret.SMTP
		groups []group.ID
		ready  bool
	}{
		{name: "nothing changed"},
		{name: "empty smtp of the system group", smtp: secret.SMTP{}, groups: []group.ID{group.System}},
		{name: "smtp of no group", smtp: secret.SMTP{Name: "private", Host: "127.0.0.1", Port: 1}},
		{name: "smtp of the system group", smtp: secret.SMTP{Name: "system", Host: "127.0.0.1", Port: 1}, groups: []group.ID{group.System}, ready: true},
	} {
		if tc.name != "nothing changed" {
			shareSMTP(t, app.Configurator(), tc.smtp, tc.groups...)
		}

		w.Click(w.Find(nagotest.ID(uionboarding.IDRecheck)))
		if !tc.ready {
			assertProblemPage(t, w)
			continue
		}

		w.FindAll(nagotest.ID(uionboarding.IDProblem)).None()
		w.FindAll(nagotest.ID(uionboarding.IDTryable)).None()
		w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
	}
}

// The Nago Mail Service is a delivery, once it is enrolled, and tryable, while a token exchange is due. After a
// failed exchange, the page shows its times.
func TestDeliveryByTheMailService(t *testing.T) {
	for _, tc := range []struct {
		name    string
		service string // "off", "available", "enrolled" or "failed"
		want    string // "problem", "tryable" or "assured"
	}{
		{name: "off", service: "off", want: "problem"},
		{name: "available", service: "available", want: "tryable"},
		{name: "enrolled", service: "enrolled", want: "assured"},
		{name: "failed exchange", service: "failed", want: "problem"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := nmstest.NewServer(t)
			switch tc.service {
			case "off":
				t.Setenv(application.EnvMailService, "off")
			case "enrolled":
				t.Setenv(application.EnvMailService, srv.URL)
				t.Setenv(application.EnvMailServiceToken, srv.IssueRefresh("tenant", "example.com"))
			default:
				t.Setenv(application.EnvMailService, srv.URL)
				t.Setenv(application.EnvMailServiceToken, "")
			}

			app, _ := pageApp(t, cfgonboarding.Options{Email: address}, nil)
			svc := std.Must(app.Configurator().MailManagement()).MailService
			if tc.service == "failed" {
				srv.VerifyOrigin = func(origin, nonce string) bool { return false }
				_, err := svc.Send(context.Background(), nms.Message{From: nms.Address{Email: "noreply@example.com"}, To: []nms.Address{{Email: "anna@example.com"}}})
				if !errors.Is(err, nms.ErrNotEnrolled) {
					t.Fatalf("expected a failed exchange, got %v", err)
				}
			}

			w := app.Open(t, nil, ".")
			switch tc.want {
			case "problem":
				assertProblemPage(t, w)
				if tc.service != "failed" {
					w.FindAll(nagotest.ID(uionboarding.IDExchangeFailed)).None()
					return
				}

				st := svc.Status()
				want := "Letzter Anmeldeversuch beim Nago Mail Service gescheitert: " + clock(st.ExchangeFailedAt) + " Uhr. Nächster Versuch frühestens " + clock(st.NextExchangeAt) + " Uhr."
				if got := w.Find(nagotest.ID(uionboarding.IDExchangeFailed)).Text(); got != want {
					t.Fatalf("expected %q, got %q", want, got)
				}
			case "tryable":
				w.Find(nagotest.ID(uionboarding.IDTryable)).Find(nagotest.Text("Diese Anwendung meldet sich gerade beim Nago Mail Service an. Ob der Code ankommt, zeigt sich erst nach dem Senden.")).Exactly(1)
				w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
			case "assured":
				w.FindAll(nagotest.ID(uionboarding.IDTryable)).None()
				w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
			}
		})
	}
}

// Without a delivery, a code is refused and no mail enters the queue. The same request succeeds with an SMTP server
// of the system group, and the code step shows the status of the mail.
func TestNoCodeMailWithoutDelivery(t *testing.T) {
	t.Setenv(application.EnvMailService, "off")
	sid := strings.Repeat("c", 32)
	app, management := pageApp(t, cfgonboarding.Options{Email: address}, nil)
	mails := std.Must(app.Configurator().MailManagement())
	queued := func() []mail.Outgoing {
		var res []mail.Outgoing
		for id, err := range mails.UseCases.FindOutgoingIDs(user.SU(), mail.OutgoingFilter{}) {
			if err != nil {
				t.Fatal(err)
			}

			res = append(res, std.Must(mails.UseCases.FindOutgoingByID(user.SU(), id)).Unwrap())
		}

		return res
	}

	if _, err := management.UseCases.RequestCode(session.ID(sid)); !errors.Is(err, onboarding.ErrNoDelivery) {
		t.Fatalf("expected no delivery, got %v", err)
	}

	if got := queued(); len(got) != 0 {
		t.Fatalf("no mail must be queued: %v", got)
	}

	shareSMTP(t, app.Configurator(), secret.SMTP{Name: "system", Host: "127.0.0.1", Port: 1}, group.System)
	if _, err := management.UseCases.RequestCode(session.ID(sid)); err != nil {
		t.Fatalf("the code must be sent: %v", err)
	}

	if got := queued(); len(got) != 1 || !strings.Contains(strings.ToLower(got[0].Receiver), address) {
		t.Fatalf("expected one mail to the address, got %v", got)
	}

	// the scheduler may have tried the unreachable server already
	w := app.Open(t, nil, ".", nagotest.Session(sid))
	w.Find(nagotest.ID("onboarding-code")).Exactly(1)
	status := w.Find(nagotest.ID(uionboarding.IDCodeStatus)).Text()
	if status != "Code in der Warteschlange, noch nicht zugestellt" && status != "Versand gescheitert, wird wiederholt" {
		t.Fatalf("unexpected status %q", status)
	}
}

// The code step shows the status of the mail with the code, nothing for an unknown status.
func TestCodeStatusLine(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status onboarding.CodeStatus
		none   bool // no CodeStatus is configured
		want   string
	}{
		{name: "no status", none: true},
		{name: "unknown", status: onboarding.CodeStatusUnknown},
		{name: "queued", status: onboarding.CodeStatusQueued, want: "Code in der Warteschlange, noch nicht zugestellt"},
		{name: "accepted", status: onboarding.CodeStatusAccepted, want: "Code vom Versandweg angenommen"},
		{name: "retrying", status: onboarding.CodeStatusRetrying, want: "Versand gescheitert, wird wiederholt"},
		{name: "failed", status: onboarding.CodeStatusFailed, want: "Code konnte nicht gesendet werden"},
		{name: "in the log", status: onboarding.CodeStatusInLog, want: "Code steht im Log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asked []string
			var mutex sync.Mutex
			opts := cfgonboarding.Options{Email: address, Deliver: (&mailbox{}).deliver}
			if !tc.none {
				opts.CodeStatus = func(to user.Email, sentAt time.Time) onboarding.CodeStatus {
					mutex.Lock()
					defer mutex.Unlock()
					asked = append(asked, fmt.Sprintf("%s %d", to, sentAt.UnixNano()))
					return tc.status
				}
			}

			app, management := pageApp(t, opts, nil)
			w := app.Open(t, nil, ".")
			w.FindAll(nagotest.ID(uionboarding.IDCodeStatus)).None() // no code yet
			w.Click(w.Find(nagotest.Text("Einrichtung starten")))
			w.Find(nagotest.ID("onboarding-code")).Exactly(1)

			if tc.want == "" {
				w.FindAll(nagotest.ID(uionboarding.IDCodeStatus)).None()
			} else if got := w.Find(nagotest.ID(uionboarding.IDCodeStatus)).Text(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}

			if tc.none {
				return
			}

			code, _ := management.UseCases.CurrentCode()
			mutex.Lock()
			defer mutex.Unlock()
			if want := fmt.Sprintf("%s %d", address, code.SentAt.UnixNano()); len(asked) == 0 || asked[len(asked)-1] != want {
				t.Fatalf("expected the status of %q, got %v", want, asked)
			}
		})
	}
}

// The delivery applies only before a code has been sent: a window in the code step stays there.
func TestStepOrder(t *testing.T) {
	delivery := level(onboarding.DeliveryNone)
	box := &mailbox{}
	sid := strings.Repeat("d", 32)
	app, management := pageApp(t, cfgonboarding.Options{
		Email:      address,
		Deliver:    box.deliver,
		Delivery:   delivery.check,
		DisableSSO: true,
		Page: func(wnd core.Window, flow *uionboarding.Flow) core.View {
			return ui.Text(fmt.Sprintf("step %d", flow.Step()))
		},
	}, nil)

	w := app.Open(t, nil, ".", nagotest.Session(sid))
	for _, tc := range []struct {
		name  string
		act   func()
		level onboarding.DeliveryLevel
		want  uionboarding.Step
	}{
		{name: "no delivery", level: onboarding.DeliveryNone, want: uionboarding.StepProblem},
		{name: "tryable", level: onboarding.DeliveryTryable, want: uionboarding.StepWelcome},
		{name: "assured", level: onboarding.DeliveryAssured, want: uionboarding.StepWelcome},
		{name: "code requested", level: onboarding.DeliveryAssured, want: uionboarding.StepCode, act: func() {
			if _, err := management.UseCases.RequestCode(session.ID(sid)); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "code requested, then no delivery", level: onboarding.DeliveryNone, want: uionboarding.StepCode},
		{name: "confirmed without delivery", level: onboarding.DeliveryNone, want: uionboarding.StepProfile, act: func() {
			if err := management.UseCases.VerifyCode(session.ID(sid), box.last(t)); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		delivery.set(onboarding.Delivery{Level: tc.level})
		if tc.act != nil {
			tc.act()
		}

		w.Reload()
		w.Find(nagotest.Text(fmt.Sprintf("step %d", tc.want))).Exactly(1)
	}

	// an invalid address is a problem, even with a delivery
	app, _ = pageApp(t, cfgonboarding.Options{Email: "no-mail", Deliver: box.deliver, Page: func(wnd core.Window, flow *uionboarding.Flow) core.View {
		return ui.Text(fmt.Sprintf("step %d", flow.Step()))
	}}, nil)
	app.Open(t, nil, ".").Find(nagotest.Text(fmt.Sprintf("step %d", uionboarding.StepProblem))).Exactly(1)
}

// storeTheme stores the legal texts of the theme before the setup starts.
func storeTheme(t *testing.T, ts theme.Settings) func(cfg *application.Configurator) {
	return func(cfg *application.Configurator) {
		sets := std.Must(cfg.SettingsManagement())
		if err := sets.UseCases.StoreGlobal(user.SU(), ts); err != nil {
			t.Fatal(err)
		}
	}
}

type link struct{ text, url, target string }

func legalLinksOf(t *testing.T, w *nagotest.Window) []link {
	t.Helper()
	var res []link
	region := w.FindAll(nagotest.ID(uionboarding.IDLegal))
	if region.Len() == 0 {
		return nil
	}

	for _, n := range region.Exactly(1).FindAll(nagotest.Type[*proto.TextView]()).Nodes() {
		tv := n.Component.(*proto.TextView)
		if tv.Link.Url == "" {
			continue // the separator
		}

		res = append(res, link{string(tv.Value), string(tv.Link.Url), string(tv.Link.Target)})
	}

	return res
}

// Below the card are the links to the legal texts of the options or the theme, each one only with a value.
func TestLegalLinks(t *testing.T) {
	for _, tc := range []struct {
		name  string
		theme theme.Settings
		opts  cfgonboarding.Options
		want  []link
	}{
		{name: "none"},
		{name: "external of the theme", theme: theme.Settings{Impress: "https://example.org/impressum", PrivacyPolicy: "https://example.org/datenschutz"},
			want: []link{{"Impressum", "https://example.org/impressum", "_blank"}, {"Datenschutz", "https://example.org/datenschutz", "_blank"}}},
		{name: "internal of the theme", theme: theme.Settings{Impress: "/impressum", PrivacyPolicy: " datenschutz "},
			want: []link{{"Impressum", "/impressum", ""}, {"Datenschutz", "/datenschutz", ""}}},
		{name: "only privacy policy", theme: theme.Settings{PrivacyPolicy: "https://example.org/datenschutz"},
			want: []link{{"Datenschutz", "https://example.org/datenschutz", "_blank"}}},
		{name: "option replaces the theme", theme: theme.Settings{Impress: "/impressum", PrivacyPolicy: "https://example.org/datenschutz"},
			opts: cfgonboarding.Options{Legal: &uionboarding.Legal{Impress: "https://example.org/imprint", PrivacyPolicy: "page/agb"}},
			want: []link{{"Impressum", "https://example.org/imprint", "_blank"}, {"Datenschutz", "/page/agb", ""}}},
		{name: "option with an empty value replaces the theme", theme: theme.Settings{Impress: "/impressum", PrivacyPolicy: "https://example.org/datenschutz"},
			opts: cfgonboarding.Options{Legal: &uionboarding.Legal{Impress: "https://example.org/imprint"}},
			want: []link{{"Impressum", "https://example.org/imprint", "_blank"}}},
		{name: "empty option replaces the theme", theme: theme.Settings{Impress: "/impressum", PrivacyPolicy: "https://example.org/datenschutz"},
			opts: cfgonboarding.Options{Legal: &uionboarding.Legal{}}},
		{name: "option without theme", opts: cfgonboarding.Options{Legal: &uionboarding.Legal{PrivacyPolicy: " /datenschutz "}},
			want: []link{{"Datenschutz", "/datenschutz", ""}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			delivery := level(onboarding.DeliveryNone)
			opts := tc.opts
			opts.Email = address
			opts.Deliver = (&mailbox{}).deliver
			opts.Delivery = delivery.check
			app, _ := pageApp(t, opts, storeTheme(t, tc.theme))
			w := app.Open(t, nil, ".")

			// on the problem, welcome and code step
			for _, step := range []string{"problem", "welcome", "code"} {
				switch step {
				case "problem":
					assertProblemPage(t, w)
				case "welcome":
					delivery.set(onboarding.Delivery{Level: onboarding.DeliveryAssured})
					w.Reload()
					w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
				case "code":
					w.Click(w.Find(nagotest.Text("Einrichtung starten")))
					w.Find(nagotest.ID("onboarding-code")).Exactly(1)
				}

				got := legalLinksOf(t, w)
				if fmt.Sprint(got) != fmt.Sprint(tc.want) {
					t.Fatalf("%s: expected %v, got %v", step, tc.want, got)
				}
			}
		})
	}
}

// The internal routes of the legal texts are shown, while the setup is pending, an external address is no route.
func TestExemptLegalRoutes(t *testing.T) {
	app, _ := pageApp(t, cfgonboarding.Options{
		Email:   address,
		Deliver: (&mailbox{}).deliver,
		Legal:   &uionboarding.Legal{PrivacyPolicy: "/datenschutz"},
		Exempt:  []core.NavigationPath{"extra"},
	}, storeTheme(t, theme.Settings{
		Impress:                   " /impressum ",
		PrivacyPolicy:             "https://example.org/other",
		GeneralTermsAndConditions: "page/agb",
		TermsOfUse:                "http://localhost/other",
	}))

	for _, tc := range []struct {
		route  core.NavigationPath
		exempt bool
	}{
		{"impressum", true},
		{"datenschutz", true},
		{"page/agb", true},
		{"extra", true},
		{"other", false},
		{".", false},
	} {
		w := app.Open(t, nil, tc.route)
		if tc.exempt {
			w.Find(nagotest.Text("page " + string(tc.route))).Exactly(1)
			w.FindAll(nagotest.Text("Einrichtung starten")).None()
			continue
		}

		w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
		w.FindAll(nagotest.TextContains("page ")).None()
	}
}

// No state of the page before the sign-in shows the address in clear, the accounts of the mail servers, a token or
// the code.
func TestPageDisclosesNothing(t *testing.T) {
	secretName, secretUser := "Geheimer Mailserver", "geheim-user"
	sixDigits := regexp.MustCompile(`\b\d{6}\b`)
	for _, tc := range []struct {
		name    string
		service string // "enrolled", "available" or "failed"
		smtp    bool   // an SMTP server of the system group
		code    bool   // request a code
		check   func(w *nagotest.Window)
	}{
		{name: "welcome, assured", service: "enrolled", smtp: true, check: func(w *nagotest.Window) {
			w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
		}},
		{name: "welcome, tryable", service: "available", check: func(w *nagotest.Window) {
			w.Find(nagotest.ID(uionboarding.IDTryable)).Exactly(1)
		}},
		{name: "code step", service: "enrolled", smtp: true, code: true, check: func(w *nagotest.Window) {
			w.Find(nagotest.ID(uionboarding.IDCodeStatus)).Exactly(1)
		}},
		{name: "not possible after a failed exchange", service: "failed", check: func(w *nagotest.Window) {
			w.Find(nagotest.ID(uionboarding.IDExchangeFailed)).Exactly(1)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := nmstest.NewServer(t)
			token := srv.IssueRefresh("tenant", "example.com")
			t.Setenv(application.EnvMailService, srv.URL)
			t.Setenv(application.EnvMailServiceToken, "")
			if tc.service == "enrolled" {
				t.Setenv(application.EnvMailServiceToken, token)
			}

			app, _ := pageApp(t, cfgonboarding.Options{Email: address}, func(cfg *application.Configurator) {
				groups := []group.ID{} // a secret of no group must not show either
				if tc.smtp {
					groups = append(groups, group.System)
				}

				shareSMTP(t, cfg, secret.SMTP{Name: secretName, Host: "127.0.0.1", Port: 1, Username: secretUser, Password: "geheim-pw"}, groups...)
			})

			if tc.service == "failed" {
				srv.VerifyOrigin = func(origin, nonce string) bool { return false }
				svc := std.Must(app.Configurator().MailManagement()).MailService
				if _, err := svc.Send(context.Background(), nms.Message{From: nms.Address{Email: "noreply@example.com"}, To: []nms.Address{{Email: "anna@example.com"}}}); err == nil {
					t.Fatal("the exchange must fail")
				}
			}

			w := app.Open(t, nil, ".")
			if tc.code {
				w.Click(w.Find(nagotest.Text("Einrichtung starten")))
			}

			tc.check(w)
			text := allText(w)
			for _, secret := range []string{address, secretName, secretUser, "geheim-pw", "127.0.0.1", token, srv.URL} {
				if strings.Contains(strings.ToLower(text), strings.ToLower(secret)) {
					t.Fatalf("the page shows %q:\n%s", secret, text)
				}
			}

			if code := sixDigits.FindString(text); code != "" {
				t.Fatalf("the page shows a code %q:\n%s", code, text)
			}
		})
	}
}

// A delivery, which is lost between the render and the click, refuses the code: the welcome shows the problem step,
// the code step an error at the field and no button to send the code again. The code, which has been sent, may
// still arrive, so its field stays. With the delivery, the same clicks send the code.
func TestLostDeliveryRefusesTheCode(t *testing.T) {
	for _, tc := range []struct {
		name  string
		level onboarding.DeliveryLevel
	}{
		{name: "lost", level: onboarding.DeliveryNone},
		{name: "kept", level: onboarding.DeliveryAssured},
	} {
		t.Run(tc.name, func(t *testing.T) {
			delivery := level(onboarding.DeliveryAssured)
			box := &mailbox{}
			app, _ := pageApp(t, cfgonboarding.Options{Email: address, Deliver: box.deliver, Delivery: delivery.check, ResendCooldown: time.Nanosecond}, nil)
			w := app.Open(t, nil, ".")

			start := w.Find(nagotest.Text("Einrichtung starten"))
			delivery.set(onboarding.Delivery{Level: tc.level})
			w.Click(start)
			w.FindAll(nagotest.Type[*proto.AlertNotifications]()).None()
			if tc.level == onboarding.DeliveryNone {
				assertProblemPage(t, w)
				if n := box.count(); n != 0 {
					t.Fatalf("no code must be sent, got %d", n)
				}

				delivery.set(onboarding.Delivery{Level: onboarding.DeliveryAssured})
				w.Click(w.Find(nagotest.ID(uionboarding.IDRecheck)))
				w.Click(w.Find(nagotest.Text("Einrichtung starten")))
			}

			// the button of the tree before the delivery changed
			resend := w.Find(nagotest.Text("Code erneut senden"))
			delivery.set(onboarding.Delivery{Level: tc.level})
			w.Click(resend)
			field := w.Find(nagotest.ID("onboarding-code")).Node().Component.(*proto.TextField)
			if tc.level != onboarding.DeliveryNone {
				if n := box.count(); field.ErrorText != "" || n != 2 {
					t.Fatalf("expected the resent code, got %q and %d codes", field.ErrorText, n)
				}

				w.Find(nagotest.Text("Code erneut senden")).Exactly(1)
				return
			}

			if n := box.count(); field.ErrorText != "Diese Anwendung kann gerade keine E-Mails versenden." || n != 1 {
				t.Fatalf("expected the refused resend, got %q and %d codes", field.ErrorText, n)
			}

			// the window stays in the code step without the button, and the code, which has been sent, is accepted
			w.Reload()
			w.FindAll(nagotest.Text("Code erneut senden")).None()
			w.Type(w.Find(nagotest.ID("onboarding-code")), box.last(t))
			w.Click(w.Find(nagotest.Text("Bestätigen")))
			w.Find(nagotest.Text("Mit E-Mail und Passwort")).Exactly(1)
		})
	}
}

// The address of the first user is read from the configured variable only, which the hint to the operators names.
func TestAddressVariable(t *testing.T) {
	const own = "NAGOTEST_FIRST_ADDRESS"
	for _, tc := range []struct {
		name    string
		option  string            // EmailEnv
		env     map[string]string // the environment
		problem string            // the variable of the hint, empty for the welcome
	}{
		{name: "default", env: map[string]string{onboarding.EnvInitialUserEmail: address, own: ""}},
		{name: "default missing", env: map[string]string{onboarding.EnvInitialUserEmail: "", own: address}, problem: onboarding.EnvInitialUserEmail},
		{name: "own", option: own, env: map[string]string{onboarding.EnvInitialUserEmail: "", own: address}},
		{name: "own missing, no fallback", option: own, env: map[string]string{onboarding.EnvInitialUserEmail: address, own: ""}, problem: own},
		{name: "own invalid", option: own, env: map[string]string{onboarding.EnvInitialUserEmail: address, own: "no-mail"}, problem: own},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			app, _ := pageApp(t, cfgonboarding.Options{EmailEnv: tc.option, Deliver: (&mailbox{}).deliver}, nil)
			w := app.Open(t, nil, ".")
			if tc.problem == "" {
				w.Find(nagotest.TextContains("t***@example.com")).Exactly(1)
				w.Find(nagotest.Text("Einrichtung starten")).Exactly(1)
				return
			}

			assertProblemPageOf(t, w, tc.problem)
		})
	}
}
