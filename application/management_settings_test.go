// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

// Acceptance tests of the wiring of the settings of a person in SettingsManagement: whose account a subject has, and
// that the records go with the account on user.Deleted. They prove the wokoda decision R-DEC-USER-SETTINGS-NAGO
// together with application/settings/my_settings_test.go.

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// wiringSidebar is a settings type of a person.
type wiringSidebar struct {
	Collapsed bool `json:"collapsed"`
}

func (wiringSidebar) UserSettings() bool { return true }

var _ = enum.Variant[settings.UserSettings, wiringSidebar](enum.Rename[wiringSidebar]("nago.test.wiring.sidebar"))

// signedIn is a person with the account id; valid false is a subject which is not signed in.
type signedIn struct {
	user.Subject
	id    user.ID
	valid bool
}

func (s signedIn) ID() user.ID { return s.id }
func (s signedIn) Valid() bool { return s.valid }

func person(id user.ID) signedIn { return signedIn{Subject: user.SU(), id: id, valid: true} }

// handledBus is the event bus of nago, which tells every finished call of a subscriber on done.
type handledBus struct {
	events.EventBus
	mu   sync.Mutex
	subs int
	done chan any
}

func (b *handledBus) Subscribe(fn func(evt any), opts ...events.SubscriberOption) func() {
	b.mu.Lock()
	b.subs++
	b.mu.Unlock()
	return b.EventBus.Subscribe(func(evt any) {
		fn(evt)
		b.done <- evt
	}, opts...)
}

func newSettingsConfigurator(t *testing.T) (*Configurator, *handledBus, settings.UseCases) {
	t.Helper()
	bus := &handledBus{EventBus: events.NewEventBus(), done: make(chan any, 16)}
	c := &Configurator{
		dataDir:   t.TempDir(),
		factories: map[proto.RootViewID]func(wnd core.Window) core.View{},
		eventBus:  bus,
	}
	sm, err := c.SettingsManagement()
	if err != nil {
		t.Fatalf("SettingsManagement: %v", err)
	}
	return c, bus, sm.UseCases
}

func loadSidebar(t *testing.T, uc settings.UseCases, s permission.Auditable) wiringSidebar {
	t.Helper()
	v, err := uc.LoadMySettings(s, reflect.TypeFor[wiringSidebar]())
	if err != nil {
		t.Fatalf("LoadMySettings: %v", err)
	}
	return v.(wiringSidebar)
}

// TestSettingsManagement_AccountOfTheSubject (R-DEC-USER-SETTINGS-NAGO): a signed-in person writes and reads her own
// record, also signed in again; another person reads the zero value. A subject without an account (not signed in, the
// system subject without id, a permission subject which is no person) reads the zero value and cannot write
// (ErrNoAccount); the counter row of each is the signed-in person.
func TestSettingsManagement_AccountOfTheSubject(t *testing.T) {
	_, _, uc := newSettingsConfigurator(t)
	anna := person("anna")
	for _, c := range []struct {
		name    string
		subject permission.Auditable
		err     error
	}{
		{"counter row: anna signed in", anna, nil},
		{"not signed in", signedIn{Subject: user.SU(), id: "anna", valid: false}, settings.ErrNoAccount},
		{"system subject", user.SU(), settings.ErrNoAccount},
		{"permission subject", permission.SU(), settings.ErrNoAccount},
	} {
		if err := uc.StoreMySettings(c.subject, wiringSidebar{Collapsed: true}); !errors.Is(err, c.err) {
			t.Errorf("%s: StoreMySettings %v, expected %v", c.name, err, c.err)
		}
	}
	for _, c := range []struct {
		name    string
		subject permission.Auditable
		want    bool
	}{
		{"anna", anna, true},
		{"anna signed in again", person("anna"), true},
		{"bert", person("bert"), false},
		{"not signed in with the id of anna", signedIn{Subject: user.SU(), id: "anna", valid: false}, false},
		{"system subject", user.SU(), false},
		{"permission subject", permission.SU(), false},
	} {
		if got := loadSidebar(t, uc, c.subject); got.Collapsed != c.want {
			t.Errorf("%s: collapsed %v, expected %v", c.name, got.Collapsed, c.want)
		}
	}
}

// TestSettingsManagement_DeletedWithTheAccount (R-DEC-USER-SETTINGS-NAGO): on user.Deleted the records of exactly
// this account are gone; those of an account whose id starts with it and of another one stay. The test waits for the
// subscriber of the wiring to finish (the bus is asynchronous), not for a time.
func TestSettingsManagement_DeletedWithTheAccount(t *testing.T) {
	c, bus, uc := newSettingsConfigurator(t)
	bus.mu.Lock()
	subs := bus.subs
	bus.mu.Unlock()
	if subs != 1 {
		t.Fatalf("premise: the wiring subscribes %d times, the test waits for exactly one subscriber", subs)
	}
	for _, id := range []user.ID{"p1", "p10", "p2"} {
		if err := uc.StoreMySettings(person(id), wiringSidebar{Collapsed: true}); err != nil {
			t.Fatal(err)
		}
	}

	c.EventBus().Publish(user.Deleted{ID: "p1"})
	if evt := <-bus.done; !reflect.DeepEqual(evt, user.Deleted{ID: "p1"}) {
		t.Fatalf("premise: the subscriber handled %#v", evt)
	}

	for _, r := range []struct {
		id   user.ID
		kept bool
	}{{"p1", false}, {"p10", true}, {"p2", true}} {
		if got := loadSidebar(t, uc, person(r.id)); got.Collapsed != r.kept {
			t.Errorf("%s: collapsed %v, expected kept=%v", r.id, got.Collapsed, r.kept)
		}
	}
}
