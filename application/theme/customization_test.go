// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/ui"
)

var brand = BaseColors{Main: "#1a73e8", Interactive: "#fbbc04", Accent: "#34a853"}

type fixture struct {
	uc      UseCases
	updates chan SettingsUpdated
}

func newFixture(t *testing.T, defaults Defaults) fixture {
	t.Helper()
	var mutex sync.Mutex
	var stored settings.GlobalSettings = Settings{}
	load := func(permission.Auditable, reflect.Type) (settings.GlobalSettings, error) {
		mutex.Lock()
		defer mutex.Unlock()
		return stored, nil
	}
	store := func(_ permission.Auditable, s settings.GlobalSettings) error {
		mutex.Lock()
		defer mutex.Unlock()
		stored = s
		return nil
	}

	bus := events.NewEventBus()
	updates := make(chan SettingsUpdated, 10)
	events.SubscribeFor[SettingsUpdated](bus, func(evt SettingsUpdated) { updates <- evt })

	return fixture{uc: NewUseCases(bus, load, store, defaults), updates: updates}
}

func (f fixture) expectUpdate(t *testing.T) {
	t.Helper()
	select {
	case <-f.updates:
	case <-time.After(time.Second):
		t.Fatal("expected a SettingsUpdated event")
	}
}

// The proposal of the application applies, until the user chooses colors, and again after a reset.
func TestUserColorsWinOverTheProposal(t *testing.T) {
	proposal := ColorsFrom(brand)
	f := newFixture(t, func() Colors { return proposal })

	colors, err := f.uc.ReadColors(user.SU())
	if err != nil {
		t.Fatal(err)
	}

	if colors != proposal {
		t.Fatal("expected the proposal of the application")
	}

	custom := CustomizationOf(proposal)
	custom.Light.Base.Main = "#eb3f33"
	if err := f.uc.UpdateCustomization(user.SU(), custom); err != nil {
		t.Fatal(err)
	}
	f.expectUpdate(t)

	colors, _ = f.uc.ReadColors(user.SU())
	if colors.Light.M0 != "#eb3f33" || colors.Dark != proposal.Dark {
		t.Fatalf("expected the colors of the user, got %v", colors.Light.M0)
	}

	if err := f.uc.ResetColors(user.SU()); err != nil {
		t.Fatal(err)
	}
	f.expectUpdate(t)

	colors, _ = f.uc.ReadColors(user.SU())
	if colors != proposal {
		t.Fatal("expected the proposal after a reset")
	}
}

func TestDefaultsFallBackToNago(t *testing.T) {
	f := newFixture(t, nil)
	colors, _ := f.uc.ReadColors(user.SU())
	if colors != ColorsFrom(DefaultBaseColors) {
		t.Fatal("expected the defaults of nago")
	}
}

// Colors, which the application has set by hand, become overrides, so that nothing is lost.
func TestCustomizationOfKeepsHandMadeColors(t *testing.T) {
	colors := ColorsFrom(brand)
	if c := CustomizationOf(colors); c.Light.Base != brand || c.Dark.Base != brand || len(c.Light.Overrides) != 0 || len(c.Dark.Overrides) != 0 {
		t.Fatalf("expected the base colors without overrides, got %+v", c)
	}

	colors.Dark.M4 = "#123456"
	colors.Light.Error = "#ff0000"
	c := CustomizationOf(colors)
	if c.Dark.Overrides["M4"] != "#123456" || c.Light.Overrides["SE0"] != "#ff0000" || len(c.Dark.Overrides) != 1 || len(c.Light.Overrides) != 1 {
		t.Fatalf("expected the hand made colors as overrides, got %+v", c)
	}

	if c.Colors() != colors {
		t.Fatal("expected the customization to calculate the same colors")
	}
}

// Overrides stay, if the base colors change.
func TestOverridesSurviveChangedBaseColors(t *testing.T) {
	f := newFixture(t, nil)
	c := Customization{
		Light: SchemeCustomization{Base: brand, Overrides: map[string]ui.Color{"M1": "#fafafa"}},
		Dark:  SchemeCustomization{Base: brand},
	}
	c.Light.Base.Accent = "#ff00ff"
	if err := f.uc.UpdateCustomization(user.SU(), c); err != nil {
		t.Fatal(err)
	}

	colors, _ := f.uc.ReadColors(user.SU())
	if colors.Light.M1 != "#fafafa" || colors.Light.A0 != "#ff00ff" {
		t.Fatalf("unexpected colors %v %v", colors.Light.M1, colors.Light.A0)
	}

	read, _ := f.uc.ReadCustomization(user.SU())
	if read.Light.Overrides["M1"] != "#fafafa" || read.Light.Base.Accent != "#ff00ff" {
		t.Fatalf("expected the stored customization, got %+v", read)
	}
}

func TestInvalidCustomizationIsRejected(t *testing.T) {
	f := newFixture(t, nil)
	for _, c := range []Customization{
		{Light: SchemeCustomization{Base: BaseColors{Main: "red", Interactive: "#fff000", Accent: "#000000"}}, Dark: SchemeCustomization{Base: brand}},
		{Light: SchemeCustomization{Base: brand, Overrides: map[string]ui.Color{"nope": "#000000"}}, Dark: SchemeCustomization{Base: brand}},
		{Light: SchemeCustomization{Base: brand}, Dark: SchemeCustomization{Base: brand, Overrides: map[string]ui.Color{"M4": "I0"}}},
	} {
		if err := f.uc.UpdateCustomization(user.SU(), c); err == nil {
			t.Errorf("expected %+v to be rejected", c)
		}
	}
}

// Colors set directly discard the customization, which would not match them anymore.
func TestUpdateColorsDiscardsTheCustomization(t *testing.T) {
	f := newFixture(t, nil)
	if err := f.uc.UpdateCustomization(user.SU(), CustomizationOf(ColorsFrom(brand))); err != nil {
		t.Fatal(err)
	}

	other := BaseColors{Main: "#6750a4", Interactive: "#b3261e", Accent: "#125d56"}
	if err := f.uc.UpdateColors(user.SU(), ColorsFrom(other)); err != nil {
		t.Fatal(err)
	}

	read, _ := f.uc.ReadCustomization(user.SU())
	if read.Light.Base != other {
		t.Fatalf("expected the customization of the new colors, got %+v", read.Light.Base)
	}
}

func TestColorKeys(t *testing.T) {
	keys := ColorKeys()
	if len(keys) != reflect.TypeFor[ui.Colors]().NumField() || keys[0] != "M0" {
		t.Fatalf("unexpected keys %v", keys)
	}

	var c ui.Colors
	for _, key := range keys {
		if !SetColor(&c, key, "#010203") || GetColor(c, key) != "#010203" {
			t.Fatalf("cannot set %s", key)
		}
	}

	if SetColor(&c, "unknown", "#000000") {
		t.Fatal("expected an unknown key to fail")
	}
}
