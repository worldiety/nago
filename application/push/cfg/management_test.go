// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgpush_test

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	stdjson "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/push"
	cfgpush "go.wdy.de/nago/application/push/cfg"
	uipush "go.wdy.de/nago/application/push/ui"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/scheduler"
	cfgscheduler "go.wdy.de/nago/application/scheduler/cfg"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/webpush"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

// de returns the German text, which the windows of nagotest show.
func de(s i18n.StrHnd) string {
	bundle, ok := i18n.Default.MatchBundle(language.German)
	if !ok {
		panic("no German bundle")
	}

	return s.Get(bundle)
}

// pushService is a fake push service, which records the messages per endpoint.
type pushService struct {
	srv      *httptest.Server
	mutex    sync.Mutex
	status   map[string]int
	received map[string][][]byte
}

func newPushService(t *testing.T) *pushService {
	p := &pushService{status: map[string]int{}, received: map[string][][]byte{}}
	p.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		p.mutex.Lock()
		defer p.mutex.Unlock()

		if !strings.HasPrefix(r.Header.Get("Authorization"), "vapid t=") || r.Header.Get("Content-Encoding") != "aes128gcm" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p.received[r.URL.Path] = append(p.received[r.URL.Path], body)
		status := p.status[r.URL.Path]
		if status == 0 {
			status = http.StatusCreated
		}

		w.WriteHeader(status)
	}))
	t.Cleanup(p.srv.Close)
	return p
}

func (p *pushService) setStatus(path string, status int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.status[path] = status
}

func (p *pushService) messages(path string) [][]byte {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.received[path]
}

// browser is a fake browser, which subscribes at the fake push service.
type browser struct {
	key      *ecdh.PrivateKey
	auth     []byte
	endpoint string
	path     string
}

func newBrowser(t *testing.T, p *pushService, name string) *browser {
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	auth := make([]byte, 16)
	_, _ = rand.Read(auth)
	return &browser{key: key, auth: auth, endpoint: p.srv.URL + "/" + name, path: "/" + name}
}

func (b *browser) subscription() *proto.RetPushSubscription {
	return &proto.RetPushSubscription{
		Endpoint:  proto.Str(b.endpoint),
		P256dh:    proto.Str(base64.RawURLEncoding.EncodeToString(b.key.PublicKey().Bytes())),
		Auth:      proto.Str(base64.RawURLEncoding.EncodeToString(b.auth)),
		Origin:    "https://app.example.com",
		UserAgent: "nagotest",
	}
}

// payloads decrypts the messages, which the browser has received.
func (b *browser) payloads(t *testing.T, p *pushService) []map[string]any {
	t.Helper()
	var res []map[string]any
	for _, msg := range p.messages(b.path) {
		plain, err := webpush.Decrypt(b.key, b.auth, msg)
		if err != nil {
			t.Fatal(err)
		}

		var m map[string]any
		if err := stdjson.Unmarshal(plain, &m); err != nil {
			t.Fatal(err)
		}

		res = append(res, m)
	}

	return res
}

type world struct {
	app   *nagotest.App
	cfg   *application.Configurator
	users user.UseCases
	mgmt  cfgpush.Management
	push  *pushService
	rid   proto.RID
}

const (
	routeCard   = "push-card"
	routeToggle = "push-toggle"
	routeDirect = "push-direct"
)

func newWorld(t *testing.T, enable bool, opts cfgpush.Options) *world {
	t.Helper()
	return newWorldAt(t, "", newPushService(t), enable, opts, nil)
}

// newWorldAt is like newWorld but keeps its data in dataDir, if not empty, so that a restart can be simulated.
// configure optionally configures more of the application.
func newWorldAt(t *testing.T, dataDir string, p *pushService, enable bool, opts cfgpush.Options, configure func(cfg *application.Configurator, w *world)) *world {
	t.Helper()
	w := &world{push: p, rid: 10_000}
	if opts.AllowEndpoint == nil {
		opts.AllowEndpoint = func(u *url.URL) bool {
			return strings.HasPrefix(u.String(), w.push.srv.URL+"/")
		}
	}

	w.app = nagotest.New(t, func(cfg *application.Configurator) {
		w.cfg = cfg
		cfg.SetApplicationID("de.worldiety.pushtest")
		if dataDir != "" {
			cfg.SetDataDir(dataDir)
		}

		if err := cfg.StandardSystems(); err != nil {
			t.Fatal(err)
		}

		if enable {
			w.mgmt = std.Must(cfgpush.Enable(cfg, opts))
		}

		if configure != nil {
			configure(cfg, w)
		}

		cfg.RootView(routeCard, func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("page"), uipush.Card(wnd), alert.BannerMessages(wnd))
		})
		cfg.RootView(routeToggle, func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("page"), uipush.Toggle(wnd))
		})
		// a button of the application, which calls Subscribe directly and shows its error
		cfg.RootView(routeDirect, func(wnd core.Window) core.View {
			result := core.AutoState[string](wnd)
			return ui.VStack(
				ui.PrimaryButton(func() {
					uipush.Subscribe(wnd, func(_ uipush.State, err error) {
						result.Update(fmt.Sprintf("result: %v", err))
					})
				}).Title("direct"),
				ui.Text(result.Get()),
			)
		})
	})

	w.users = std.Must(w.cfg.UserManagement()).UseCases
	return w
}

func (w *world) person(t *testing.T, name string) user.Subject {
	t.Helper()
	mail := fmt.Sprintf("%s@example.com", strings.ToLower(name))
	u, err := w.users.Create(user.SU(), user.ShortRegistrationUser{Firstname: name, Lastname: "Test", Email: user.Email(mail), Password: "Sup3r-Geheim!2026", PasswordRepeated: "Sup3r-Geheim!2026", Verified: true})
	if err != nil {
		t.Fatal(err)
	}

	return std.Must(w.users.SubjectFromUser(user.SU(), u.ID)).Unwrap()
}

// pushCalls returns the push calls of the window by type.
func pushCalls[T proto.CallArgs](win *nagotest.Window) []*proto.CallRequested {
	var res []*proto.CallRequested
	for _, call := range win.AsyncCalls() {
		if _, ok := call.Call.(T); ok {
			res = append(res, call)
		}
	}

	return res
}

// anyPushCall reports the push calls of the window, which would let the frontend touch web push at all.
func anyPushCall(win *nagotest.Window) []string {
	var res []string
	for _, call := range win.AsyncCalls() {
		switch call.Call.(type) {
		case *proto.CallPushState, *proto.CallPushSubscribe, *proto.CallPushUnsubscribe:
			res = append(res, fmt.Sprintf("%T", call.Call))
		}
	}

	return res
}

func (w *world) resolve(t *testing.T, win *nagotest.Window, call *proto.CallRequested, ret proto.CallRet) {
	t.Helper()
	w.rid++
	if err := win.Scope().Dispatch(&proto.CallResolved{CallPtr: call.CallPtr, RID: w.rid, Ret: ret}); err != nil {
		t.Fatal(err)
	}

	win.Settle()
}

func defaultState() *proto.RetPushState {
	return &proto.RetPushState{Keep: true, Supported: true, Permission: "default"}
}

// subscribe opens the card, lets the browser tell its state and clicks the button.
func (w *world) subscribe(t *testing.T, subject user.Subject, sid string, b *browser) *nagotest.Window {
	t.Helper()
	win := w.app.Open(t, subject, routeCard, nagotest.Session(sid))
	w.resolve(t, win, pushCalls[*proto.CallPushState](win)[0], defaultState())
	win.Click(win.Find(nagotest.Text(de(uipush.StrEnable))))
	calls := pushCalls[*proto.CallPushSubscribe](win)
	if len(calls) != 1 {
		t.Fatalf("expected one subscribe call after the click, got %d", len(calls))
	}

	w.resolve(t, win, calls[0], b.subscription())
	return win
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %s", what)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func countSubscriptions(t *testing.T, w *world, subject user.Subject) int {
	t.Helper()
	n := 0
	for _, err := range w.mgmt.UseCases.FindMySubscriptions(subject) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}

	return n
}

// TestWithoutModule shows, that the views of uipush neither render nor make the frontend touch web push, if the
// module is not enabled.
func TestWithoutModule(t *testing.T) {
	w := newWorld(t, false, cfgpush.Options{})
	ada := w.person(t, "Ada")

	for _, route := range []core.NavigationPath{routeCard, routeToggle} {
		win := w.app.Open(t, ada, route)
		win.Find(nagotest.Text("page"))
		win.FindAll(nagotest.Text(de(uipush.StrTitle))).None()
		if calls := anyPushCall(win); len(calls) != 0 {
			t.Fatalf("%s: unexpected push calls without module: %v", route, calls)
		}
	}

	win := w.app.Open(t, ada, routeDirect)
	win.Click(win.Find(nagotest.Text("direct")))
	win.Find(nagotest.Text("result: " + uipush.ErrNotEnabled.Error()))
	if calls := anyPushCall(win); len(calls) != 0 {
		t.Fatalf("Subscribe without module must not call the frontend: %v", calls)
	}
}

// TestNoPermissionRequestWithoutGesture shows, that the enabled module only reads the state of the browser, until
// the user clicks the button of the card.
func TestNoPermissionRequestWithoutGesture(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")

	win := w.app.Open(t, ada, routeCard)
	if n := len(pushCalls[*proto.CallPushState](win)); n != 1 {
		t.Fatalf("expected one state call, got %d", n)
	}

	w.resolve(t, win, pushCalls[*proto.CallPushState](win)[0], defaultState())
	win.Find(nagotest.Text(de(uipush.StrInactive)))
	win.Find(nagotest.Text(de(uipush.StrEnable)))

	if n := len(pushCalls[*proto.CallPushSubscribe](win)); n != 0 {
		t.Fatalf("the browser must not be asked without a click, got %d subscribe calls", n)
	}

	// a guest has no user to get notifications for, so the card offers no button and the browser is not asked
	guestCard := w.app.Open(t, nil, routeCard)
	w.resolve(t, guestCard, pushCalls[*proto.CallPushState](guestCard)[0], defaultState())
	guestCard.Find(nagotest.Text(de(uipush.StrSignIn)))
	guestCard.FindAll(nagotest.Text(de(uipush.StrEnable))).None()

	guest := w.app.Open(t, nil, routeDirect)
	guest.Click(guest.Find(nagotest.Text("direct")))
	guest.Find(nagotest.Text("result: " + user.InvalidSubjectErr.Error()))
	if n := len(pushCalls[*proto.CallPushSubscribe](guest)); n != 0 {
		t.Fatalf("a guest must not be asked, got %d subscribe calls", n)
	}

	// a denied permission and a browser without push offer no button, which could ask
	for _, state := range []*proto.RetPushState{
		{Keep: true, Supported: true, Permission: "denied"},
		{Keep: true},
		{Keep: true, RequiresInstall: true},
	} {
		win := w.app.Open(t, ada, routeCard)
		w.resolve(t, win, pushCalls[*proto.CallPushState](win)[0], state)
		win.FindAll(nagotest.Text(de(uipush.StrEnable))).None()
	}
}

func TestSubscribeAndSend(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")
	bob := w.person(t, "Bob")
	phone := newBrowser(t, w.push, "ada-phone")

	win := w.subscribe(t, ada, "session-ada", phone)
	call := pushCalls[*proto.CallPushSubscribe](win)[0].Call.(*proto.CallPushSubscribe)
	if string(call.ApplicationServerKey) != w.mgmt.UseCases.PublicKey() || string(call.ServiceWorker) != push.ServiceWorkerPath {
		t.Fatalf("unexpected call %+v", call)
	}

	win.Find(nagotest.Text(de(uipush.StrActive)))
	if n := countSubscriptions(t, w, ada); n != 1 {
		t.Fatalf("expected one subscription, got %d", n)
	}

	// without the permission, nobody may send to others
	if _, err := w.mgmt.UseCases.Send(bob, push.ToUser(ada.ID()), push.Message{Title: "x"}); err == nil {
		t.Fatal("expected a permission error")
	}

	report, err := w.mgmt.UseCases.Send(user.SU(), push.ToUser(ada.ID(), bob.ID()), push.Message{
		Title:  "Neuer Auftrag",
		Body:   "Auftrag 42 wartet",
		Route:  "orders",
		Values: core.Values{"id": "42"},
		Tag:    "order-42",
	})
	if err != nil {
		t.Fatal(err)
	}

	if report != (push.Report{Users: 2, Sent: 1}) {
		t.Fatalf("unexpected report %+v", report)
	}

	got := phone.payloads(t, w.push)
	if len(got) != 1 || got[0]["title"] != "Neuer Auftrag" || got[0]["body"] != "Auftrag 42 wartet" || got[0]["url"] != "/orders?id=42" || got[0]["tag"] != "order-42" {
		t.Fatalf("unexpected payloads %v", got)
	}

	// the test button of the card sends to the own browsers only
	win.Click(win.Find(nagotest.Text(de(uipush.StrSendTest))))
	if got := phone.payloads(t, w.push); len(got) != 2 || got[1]["title"] != de(uipush.StrTestTitle) {
		t.Fatalf("unexpected test payload %v", got)
	}

	// a new state call tells, that the browser is registered for the user
	win2 := w.app.Open(t, ada, routeCard)
	w.resolve(t, win2, pushCalls[*proto.CallPushState](win2)[0], &proto.RetPushState{Keep: true, Supported: true, Permission: "granted", Endpoint: proto.Str(phone.endpoint)})
	win2.Find(nagotest.Text(de(uipush.StrActive)))

	// the same browser of another user is the same subscription, which changes its owner
	w.subscribe(t, bob, "session-bob", phone)
	if countSubscriptions(t, w, ada) != 0 || countSubscriptions(t, w, bob) != 1 {
		t.Fatal("the browser must belong to the last user only")
	}

	// unsubscribe removes it
	bobWin := w.app.Open(t, bob, routeCard)
	w.resolve(t, bobWin, pushCalls[*proto.CallPushState](bobWin)[0], &proto.RetPushState{Keep: true, Supported: true, Permission: "granted", Endpoint: proto.Str(phone.endpoint)})
	bobWin.Click(bobWin.Find(nagotest.Text(de(uipush.StrDisable))))
	w.resolve(t, bobWin, pushCalls[*proto.CallPushUnsubscribe](bobWin)[0], phone.subscription())
	if countSubscriptions(t, w, bob) != 0 {
		t.Fatal("unsubscribe must delete the subscription")
	}
}

func TestSendToGroupAndRole(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")
	bob := w.person(t, "Bob")
	eve := w.person(t, "Eve")

	browsers := map[string]*browser{}
	for _, p := range []struct {
		name    string
		subject user.Subject
	}{{"ada", ada}, {"bob", bob}, {"eve", eve}} {
		browsers[p.name] = newBrowser(t, w.push, p.name)
		w.subscribe(t, p.subject, "session-"+p.name, browsers[p.name])
	}

	rdb := std.Must(w.cfg.RDB())
	member := func(ns rebac.Namespace, instance string, uid user.ID) {
		if err := rdb.Put(rebac.Triple{
			Source:   rebac.Entity{Namespace: ns, Instance: rebac.Instance(instance)},
			Relation: rebac.Member,
			Target:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(uid)},
		}); err != nil {
			t.Fatal(err)
		}
	}

	member(group.Namespace, "sales", ada.ID())
	member(group.Namespace, "sales", bob.ID())
	member(role.Namespace, "dispatcher", bob.ID())
	member(role.Namespace, "dispatcher", eve.ID())

	report, err := w.mgmt.UseCases.Send(user.SU(), push.ToGroup("sales"), push.Message{Title: "Gruppe"})
	if err != nil || report.Sent != 2 || report.Users != 2 {
		t.Fatalf("unexpected report %+v %v", report, err)
	}

	// a user in the group and in the role gets the message once
	report, err = w.mgmt.UseCases.Send(user.SU(), push.Recipients{Groups: []group.ID{"sales"}, Roles: []role.ID{"dispatcher"}}, push.Message{Title: "Beide"})
	if err != nil || report.Sent != 3 || report.Users != 3 {
		t.Fatalf("unexpected report %+v %v", report, err)
	}

	count := func(name string) int { return len(browsers[name].payloads(t, w.push)) }
	if count("ada") != 2 || count("bob") != 2 || count("eve") != 1 {
		t.Fatalf("unexpected deliveries ada=%d bob=%d eve=%d", count("ada"), count("bob"), count("eve"))
	}
}

func TestCleanup(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")
	bob := w.person(t, "Bob")

	// a push service, which does not know the subscription anymore, lets it be deleted
	gone := newBrowser(t, w.push, "gone")
	w.subscribe(t, ada, "session-ada-1", gone)
	w.push.setStatus(gone.path, http.StatusGone)
	report, err := w.mgmt.UseCases.Send(user.SU(), push.ToUser(ada.ID()), push.Message{Title: "x"})
	if err != nil || report.Removed != 1 || countSubscriptions(t, w, ada) != 0 {
		t.Fatalf("410 must remove the subscription: %+v %v", report, err)
	}

	notFound := newBrowser(t, w.push, "not-found")
	w.subscribe(t, ada, "session-ada-1", notFound)
	w.push.setStatus(notFound.path, http.StatusNotFound)
	if report, _ := w.mgmt.UseCases.Send(user.SU(), push.ToUser(ada.ID()), push.Message{Title: "x"}); report.Removed != 1 {
		t.Fatalf("404 must remove the subscription: %+v", report)
	}

	// another error keeps it
	failing := newBrowser(t, w.push, "failing")
	w.subscribe(t, ada, "session-ada-1", failing)
	w.push.setStatus(failing.path, http.StatusTooManyRequests)
	if report, _ := w.mgmt.UseCases.Send(user.SU(), push.ToUser(ada.ID()), push.Message{Title: "x"}); report.Failed != 1 || countSubscriptions(t, w, ada) != 1 {
		t.Fatalf("429 must keep the subscription: %+v", report)
	}

	// the logout of a session deletes its subscriptions, but not those of other sessions
	w.subscribe(t, ada, "session-ada-2", newBrowser(t, w.push, "ada-2"))
	if countSubscriptions(t, w, ada) != 2 {
		t.Fatal("expected two subscriptions")
	}

	w.cfg.EventBus().Publish(session.LoggedOut{Session: "session-ada-1"})
	waitFor(t, "logout cleanup", func() bool { return countSubscriptions(t, w, ada) == 1 })

	// deleting the user deletes the subscriptions
	w.subscribe(t, bob, "session-bob", newBrowser(t, w.push, "bob"))
	if err := w.users.Delete(user.SU(), bob.ID()); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "user cleanup", func() bool {
		n := 0
		for sub, err := range subscriptions(t, w).All() {
			if err != nil {
				t.Fatal(err)
			}
			if sub.User == bob.ID() {
				n++
			}
		}
		return n == 0
	})
}

func TestKeepOnLogout(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{KeepOnLogout: true})
	ada := w.person(t, "Ada")
	w.subscribe(t, ada, "session-ada", newBrowser(t, w.push, "ada"))
	w.cfg.EventBus().Publish(session.LoggedOut{Session: "session-ada"})
	time.Sleep(100 * time.Millisecond)
	if countSubscriptions(t, w, ada) != 1 {
		t.Fatal("KeepOnLogout must keep the subscription")
	}
}

func TestRegisterRejectsForeignEndpoints(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{AllowEndpoint: push.KnownPushService})
	ada := w.person(t, "Ada")
	b := newBrowser(t, w.push, "internal")

	_, err := w.mgmt.UseCases.Register(ada, "s", push.BrowserSubscription{Endpoint: b.endpoint, P256dh: string(b.subscription().P256dh), Auth: string(b.subscription().Auth)})
	if err == nil {
		t.Fatal("an endpoint of an unknown host must be rejected")
	}

	for _, endpoint := range []string{
		"https://fcm.googleapis.com/fcm/send/abc",
		"https://updates.push.services.mozilla.com/wpush/v2/abc",
		"https://web.push.apple.com/abc",
		"https://wns2-par02p.notify.windows.com/w/?token=abc",
	} {
		if _, err := w.mgmt.UseCases.Register(ada, "s", push.BrowserSubscription{Endpoint: endpoint, P256dh: string(b.subscription().P256dh), Auth: string(b.subscription().Auth)}); err != nil {
			t.Fatalf("%s must be allowed: %v", endpoint, err)
		}
	}

	for _, endpoint := range []string{
		"http://fcm.googleapis.com/fcm/send/abc",
		"https://fcm.googleapis.com.evil.example/abc",
		"https://fcm.googleapis.com:8443/abc",
		"https://user@fcm.googleapis.com/abc",
	} {
		if _, err := w.mgmt.UseCases.Register(ada, "s", push.BrowserSubscription{Endpoint: endpoint, P256dh: string(b.subscription().P256dh), Auth: string(b.subscription().Auth)}); err == nil {
			t.Fatalf("%s must be rejected", endpoint)
		}
	}

	if _, err := w.mgmt.UseCases.Register(ada, "s", push.BrowserSubscription{Endpoint: "https://fcm.googleapis.com/x", P256dh: "AAAA", Auth: "AAAA"}); err == nil {
		t.Fatal("invalid keys must be rejected")
	}
}

func TestServiceWorkerRoute(t *testing.T) {
	get := func(t *testing.T, enable bool) *http.Response {
		base := nagotest.Serve(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID("de.worldiety.pushtest.http")
			if err := cfg.StandardSystems(); err != nil {
				t.Fatal(err)
			}

			if enable {
				std.Must(cfgpush.Enable(cfg, cfgpush.Options{}))
			}
		})

		res, err := http.Get(base + push.ServiceWorkerPath)
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = res.Body.Close() })
		return res
	}

	res := get(t, true)
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK || res.Header.Get("Service-Worker-Allowed") != "/" || !strings.HasPrefix(res.Header.Get("Content-Type"), "text/javascript") || !strings.Contains(string(body), "notificationclick") {
		t.Fatalf("unexpected service worker response %d %v", res.StatusCode, res.Header)
	}

	if strings.Contains(string(body), "addEventListener('fetch'") {
		t.Fatal("the service worker must not intercept requests")
	}

	res = get(t, false)
	body, _ = io.ReadAll(res.Body)
	if res.Header.Get("Service-Worker-Allowed") != "" || strings.Contains(string(body), "notificationclick") {
		t.Fatal("without the module, there must be no service worker")
	}
}

// subscriptions returns the stored subscriptions, which only the master key can read.
func subscriptions(t *testing.T, w *world) data.Repository[push.Subscription, push.SubscriptionID] {
	t.Helper()
	store := std.Must(w.cfg.EntityStore("nago.push.subscription"))
	return json.NewSloppyJSONRepository[push.Subscription, push.SubscriptionID](crypto.NewBlobStore(store, std.Must(w.cfg.MasterKey())))
}

// raw returns the stored bytes of a bucket, as they are on disk.
func raw(t *testing.T, w *world, bucket string) []byte {
	t.Helper()
	store := std.Must(w.cfg.EntityStore(bucket))
	var all []byte
	for _, key := range std.Must(blob.Keys(store)) {
		opt := std.Must(blob.Get(store, key))
		all = append(all, opt.Unwrap()...)
	}

	return all
}

// TestStoredEncrypted shows, that neither the subscriptions nor the VAPID key are readable without the master key.
func TestStoredEncrypted(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")
	phone := newBrowser(t, w.push, "ada-phone")
	w.subscribe(t, ada, "session-ada", phone)

	stored := raw(t, w, "nago.push.subscription")
	if len(stored) == 0 {
		t.Fatal("expected a stored subscription")
	}

	for _, secret := range []string{phone.endpoint, string(phone.subscription().Auth), string(phone.subscription().P256dh), string(ada.ID()), "endpoint"} {
		if bytes.Contains(stored, []byte(secret)) {
			t.Fatalf("the subscription is stored in plain text: %q", secret)
		}
	}

	if sub := std.Must(subscriptions(t, w).FindByID(push.NewSubscriptionID(phone.endpoint))); sub.IsNone() || sub.Unwrap().Endpoint != phone.endpoint {
		t.Fatal("the subscription must be readable with the master key")
	}

	key := raw(t, w, "nago.push.vapid")
	if len(key) == 0 || bytes.Contains(key, []byte("privateKey")) {
		t.Fatal("the VAPID key must be stored encrypted")
	}
}

// TestSchedulerAfterRestart shows, that a scheduler sends as the system user and that the subscription and the VAPID
// key survive a restart, which keeps the data directory. The browser keeps its subscription anyway.
func TestSchedulerAfterRestart(t *testing.T) {
	dataDir := t.TempDir()
	p := newPushService(t)
	phone := newBrowser(t, p, "ada-phone")

	// the job of the application, e.g. a reminder of the open orders of each morning
	reminders := func(cfg *application.Configurator, w *world) {
		schedulers := std.Must(cfgscheduler.Enable(cfg))
		if err := schedulers.UseCases.Configure(user.SU(), scheduler.Options{
			ID:   "test.push.reminder",
			Name: "Erinnerung",
			Kind: scheduler.Manual,
			Runner: func(ctx context.Context) error {
				report, err := w.mgmt.UseCases.Send(user.SU(), push.ToGroup("sales"), push.Message{
					Title: "Offene Aufträge",
					Route: "orders",
				})
				if err != nil {
					return err
				}

				if report.Failed > 0 {
					return fmt.Errorf("push failed for %d devices", report.Failed)
				}

				return nil
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	var publicKey string
	var ada user.ID
	t.Run("before", func(t *testing.T) {
		w := newWorldAt(t, dataDir, p, true, cfgpush.Options{}, reminders)
		subject := w.person(t, "Ada")
		ada = subject.ID()
		w.subscribe(t, subject, "session-ada", phone)
		publicKey = w.mgmt.UseCases.PublicKey()
	})

	// the first application has been destroyed at the end of the subtest
	w := newWorldAt(t, dataDir, p, true, cfgpush.Options{}, reminders)
	if w.mgmt.UseCases.PublicKey() != publicKey {
		t.Fatal("the VAPID key must survive the restart, otherwise all browsers would have to subscribe again")
	}

	rdb := std.Must(w.cfg.RDB())
	if err := rdb.Put(rebac.Triple{
		Source:   rebac.Entity{Namespace: group.Namespace, Instance: "sales"},
		Relation: rebac.Member,
		Target:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(ada)},
	}); err != nil {
		t.Fatal(err)
	}

	schedulers := std.Must(cfgscheduler.Enable(w.cfg))
	if err := schedulers.UseCases.ExecuteNow(user.SU(), "test.push.reminder"); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "the message of the scheduler", func() bool { return len(p.messages(phone.path)) == 1 })
	if got := phone.payloads(t, p); got[0]["title"] != "Offene Aufträge" || got[0]["url"] != "/orders" {
		t.Fatalf("unexpected payload %v", got)
	}
}

// TestFailedSubscription shows, that an error of the browser keeps the last known state, so the user may try again.
func TestFailedSubscription(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")

	win := w.app.Open(t, ada, routeCard)
	w.resolve(t, win, pushCalls[*proto.CallPushState](win)[0], defaultState())
	win.Click(win.Find(nagotest.Text(de(uipush.StrEnable))))
	w.resolve(t, win, pushCalls[*proto.CallPushSubscribe](win)[0], &proto.RetError{Message: "push service not available", Code: 500})
	win.Find(nagotest.Text(de(uipush.StrInactive)))
	win.Find(nagotest.Text(de(uipush.StrEnable)))
	if countSubscriptions(t, w, ada) != 0 {
		t.Fatal("a failed subscription must not be stored")
	}
}

// TestSubscriptionWithoutAnswer shows, that a push service, which does not answer, tells the user so.
func TestSubscriptionWithoutAnswer(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")

	win := w.app.Open(t, ada, routeCard)
	w.resolve(t, win, pushCalls[*proto.CallPushState](win)[0], defaultState())
	win.Click(win.Find(nagotest.Text(de(uipush.StrEnable))))
	w.resolve(t, win, pushCalls[*proto.CallPushSubscribe](win)[0], &proto.RetError{Message: "the push service has not answered within 30000 ms", Code: 504})
	win.Find(nagotest.TextContains(de(uipush.StrNoAnswer)))
	win.Find(nagotest.Text(de(uipush.StrEnable)))

	// the error of Subscribe matches ErrNoAnswer
	direct := w.app.Open(t, ada, routeDirect)
	direct.Click(direct.Find(nagotest.Text("direct")))
	w.resolve(t, direct, pushCalls[*proto.CallPushSubscribe](direct)[0], &proto.RetError{Message: "timeout", Code: 504})
	direct.Find(nagotest.TextContains(uipush.ErrNoAnswer.Error()))
}

func TestToggle(t *testing.T) {
	w := newWorld(t, true, cfgpush.Options{})
	ada := w.person(t, "Ada")
	b := newBrowser(t, w.push, "ada")

	win := w.app.Open(t, ada, routeToggle)
	w.resolve(t, win, pushCalls[*proto.CallPushState](win)[0], defaultState())
	if n := len(pushCalls[*proto.CallPushSubscribe](win)); n != 0 {
		t.Fatalf("no subscribe without switching, got %d", n)
	}

	win.Type(win.Find(nagotest.Type[*proto.Toggle]()), "true")
	calls := pushCalls[*proto.CallPushSubscribe](win)
	if len(calls) != 1 {
		t.Fatalf("switching on must subscribe once, got %d", len(calls))
	}

	w.resolve(t, win, calls[0], b.subscription())
	if countSubscriptions(t, w, ada) != 1 {
		t.Fatal("expected a subscription")
	}
}
