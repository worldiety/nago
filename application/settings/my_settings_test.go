// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package settings_test

// Acceptance tests of the settings of a person (LoadMySettings, StoreMySettings, DeleteAccount). They prove the
// wokoda decision R-DEC-USER-SETTINGS-NAGO: one record per account and type, the zero value without a record, read
// and written only for the own account, kept over a restart, deleted with the account. The account of a subject comes
// from the AccountOf of the test here; the wiring of package application is tested in
// application/management_settings_test.go.

import (
	"errors"
	"reflect"
	"testing"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
)

// testSidebar and testRights are two settings types of a person; their zero values are the defaults.
type testSidebar struct {
	Collapsed bool `json:"collapsed"`
}

func (testSidebar) UserSettings() bool { return true }

type testRights struct {
	Off bool `json:"off"`
}

func (testRights) UserSettings() bool { return true }

var (
	_ = enum.Variant[settings.UserSettings, testSidebar](enum.Rename[testSidebar]("nago.test.settings.sidebar"))
	_ = enum.Variant[settings.UserSettings, testRights](enum.Rename[testRights]("nago.test.settings.rights"))
)

// person is a subject with an account, or without one for an empty account.
type person struct{ account string }

func (person) Audit(permission.ID) error        { return errors.New("denied") }
func (person) HasPermission(permission.ID) bool { return false }

func accountOf(s permission.Auditable) (string, bool) {
	p, ok := s.(person)
	if !ok || p.account == "" {
		return "", false
	}
	return p.account, true
}

type stores struct {
	global data.Repository[settings.StoreBox[settings.GlobalSettings], settings.ID]
	user   data.Repository[settings.StoreBox[settings.UserSettings], settings.ID]
}

func newStores() stores {
	return stores{
		global: json.NewSloppyJSONRepository[settings.StoreBox[settings.GlobalSettings]](mem.NewBlobStore("global")),
		user:   json.NewSloppyJSONRepository[settings.StoreBox[settings.UserSettings]](mem.NewBlobStore("user")),
	}
}

func (s stores) useCases() settings.UseCases {
	return settings.NewUseCases(s.global, s.user, events.NewEventBus(), accountOf)
}

func load[T settings.UserSettings](t *testing.T, uc settings.UseCases, s permission.Auditable) T {
	t.Helper()
	v, err := uc.LoadMySettings(s, reflect.TypeFor[T]())
	if err != nil {
		t.Fatalf("LoadMySettings: %v", err)
	}
	return v.(T)
}

// TestMySettings_OwnAccountOnly (R-DEC-USER-SETTINGS-NAGO): a person reads what she wrote, also as a new subject of
// the same account (signed in again); another person reads the zero value and her own write does not touch the first;
// the two types of one account are two records. Without an account, Load gives the zero value and Store fails with
// ErrNoAccount; its counter row is the same write with an account.
func TestMySettings_OwnAccountOnly(t *testing.T) {
	uc := newStores().useCases()
	anna, bert, nobody := person{"anna"}, person{"bert"}, person{}

	if got := load[testSidebar](t, uc, anna); got != (testSidebar{}) {
		t.Fatalf("premise: without a record %+v, expected the zero value", got)
	}

	for _, c := range []struct {
		name    string
		subject permission.Auditable
		write   settings.UserSettings
		err     error
	}{
		{"anna writes her sidebar", anna, testSidebar{Collapsed: true}, nil},
		{"without an account", nobody, testSidebar{Collapsed: true}, settings.ErrNoAccount},
		{"not a person at all", permission.SU(), testSidebar{Collapsed: true}, settings.ErrNoAccount},
	} {
		if err := uc.StoreMySettings(c.subject, c.write); !errors.Is(err, c.err) {
			t.Errorf("%s: StoreMySettings %v, expected %v", c.name, err, c.err)
		}
	}

	for _, c := range []struct {
		name    string
		subject permission.Auditable
		want    testSidebar
	}{
		{"anna reads it", anna, testSidebar{Collapsed: true}},
		{"anna signed in again", person{"anna"}, testSidebar{Collapsed: true}},
		{"bert reads the zero value", bert, testSidebar{}},
		{"without an account the zero value", nobody, testSidebar{}},
		{"not a person at all, the zero value", permission.SU(), testSidebar{}},
	} {
		if got := load[testSidebar](t, uc, c.subject); got != c.want {
			t.Errorf("%s: %+v, expected %+v", c.name, got, c.want)
		}
	}

	// bert writes only his own record
	if err := uc.StoreMySettings(bert, testSidebar{Collapsed: false}); err != nil {
		t.Fatal(err)
	}
	if err := uc.StoreMySettings(bert, testRights{Off: true}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name    string
		subject permission.Auditable
		sidebar testSidebar
		rights  testRights
	}{
		{"anna after bert wrote", anna, testSidebar{Collapsed: true}, testRights{}},
		{"bert, two types", bert, testSidebar{}, testRights{Off: true}},
	} {
		if got := load[testSidebar](t, uc, c.subject); got != c.sidebar {
			t.Errorf("%s: sidebar %+v, expected %+v", c.name, got, c.sidebar)
		}
		if got := load[testRights](t, uc, c.subject); got != c.rights {
			t.Errorf("%s: rights %+v, expected %+v", c.name, got, c.rights)
		}
	}

	// MySettings is the way an application reads it
	if got := settings.MySettings[testSidebar](anna, uc.LoadMySettings); !got.Collapsed {
		t.Errorf("MySettings for anna: %+v, expected collapsed", got)
	}
}

// TestMySettings_KeptOverRestart (R-DEC-USER-SETTINGS-NAGO): new use cases on the same stores, as after a restart,
// read what the old ones wrote; counter row: new use cases on new stores read the zero value.
func TestMySettings_KeptOverRestart(t *testing.T) {
	st := newStores()
	anna := person{"anna"}
	if err := st.useCases().StoreMySettings(anna, testSidebar{Collapsed: true}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		uc   settings.UseCases
		want testSidebar
	}{
		{"restart on the same stores", st.useCases(), testSidebar{Collapsed: true}},
		{"counter row: new stores", newStores().useCases(), testSidebar{}},
	} {
		if got := load[testSidebar](t, c.uc, anna); got != c.want {
			t.Errorf("%s: %+v, expected %+v", c.name, got, c.want)
		}
	}
}

// TestMySettings_DeleteAccount (R-DEC-USER-SETTINGS-NAGO): DeleteAccount removes every record of the account, of both
// types, and nothing of another account, also not of one whose name starts with the deleted one.
func TestMySettings_DeleteAccount(t *testing.T) {
	st := newStores()
	uc := st.useCases()
	for _, a := range []string{"p1", "p10", "p2"} {
		if err := uc.StoreMySettings(person{a}, testSidebar{Collapsed: true}); err != nil {
			t.Fatal(err)
		}
		if err := uc.StoreMySettings(person{a}, testRights{Off: true}); err != nil {
			t.Fatal(err)
		}
	}
	if err := settings.DeleteAccount(st.user, "p1"); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		account string
		kept    bool
	}{{"p1", false}, {"p10", true}, {"p2", true}} {
		s, r := load[testSidebar](t, uc, person{c.account}), load[testRights](t, uc, person{c.account})
		if s.Collapsed != c.kept || r.Off != c.kept {
			t.Errorf("%s: sidebar %+v, rights %+v, expected kept=%v", c.account, s, r, c.kept)
		}
	}
}
