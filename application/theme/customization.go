// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"fmt"
	"reflect"
	"strings"

	"go.wdy.de/nago/pkg/xcolor"
	"go.wdy.de/nago/presentation/ui"
)

// Customization keeps what a user has chosen for the theme colors, so that it can be edited again. The colors of
// the theme are calculated from it, see [Customization.Colors].
type Customization struct {
	Light SchemeCustomization `json:"light"`
	Dark  SchemeCustomization `json:"dark"`
}

// SchemeCustomization keeps the choice for either the light or the dark mode.
type SchemeCustomization struct {
	// Base are the three colors, from which all other colors are calculated.
	Base BaseColors `json:"base"`
	// Overrides replace single calculated colors. The key is the json name of the [ui.Colors] field, e.g. M4.
	Overrides map[string]ui.Color `json:"overrides,omitempty"`
}

// Valid tells if the base colors of both modes are set.
func (c Customization) Valid() bool {
	return c.Light.Base.valid() && c.Dark.Base.valid()
}

// Colors calculates the colors of the theme.
func (c Customization) Colors() Colors {
	return Colors{
		Light: c.Light.Colors(LightMode),
		Dark:  c.Dark.Colors(DarkMode),
	}
}

// Colors calculates the colors of the mode from the base colors and replaces the overridden ones.
func (c SchemeCustomization) Colors(derive DeriveColors) ui.Colors {
	colors := derive(c.Base)
	for key, color := range c.Overrides {
		SetColor(&colors, key, color)
	}

	return colors
}

func (c SchemeCustomization) validate() error {
	if !c.Base.valid() {
		return fmt.Errorf("invalid base colors: %v", c.Base)
	}

	for key, color := range c.Overrides {
		if _, ok := colorFieldIndex(key); !ok {
			return fmt.Errorf("unknown color %q", key)
		}

		if !isHex(color) {
			return fmt.Errorf("invalid color %q for %s", color, key)
		}
	}

	return nil
}

func (b BaseColors) valid() bool {
	return isHex(b.Main) && isHex(b.Interactive) && isHex(b.Accent)
}

func isHex(c ui.Color) bool {
	_, err := xcolor.ParseHex(string(c))
	return err == nil
}

// ColorsFrom calculates the colors of both modes from the given base colors. Use it to propose the colors of an
// application, see [go.wdy.de/nago/application.Configurator.SetThemeDefaults].
func ColorsFrom(base BaseColors) Colors {
	return Colors{
		Light: LightMode(base),
		Dark:  DarkMode(base),
	}
}

// CustomizationOf returns the customization, which calculates the given colors. The base colors are taken from
// M0, I0 and A0, and every color, which differs from the calculated one, becomes an override. Thus, also colors
// which have not been calculated by nago can be edited without loss.
func CustomizationOf(colors Colors) Customization {
	return Customization{
		Light: schemeCustomizationOf(colors.Light, LightMode),
		Dark:  schemeCustomizationOf(colors.Dark, DarkMode),
	}
}

func schemeCustomizationOf(colors ui.Colors, derive DeriveColors) SchemeCustomization {
	c := SchemeCustomization{Base: BaseColors{Main: colors.M0, Interactive: colors.I0, Accent: colors.A0}}
	if !c.Base.valid() {
		c.Base = DefaultBaseColors
	}

	calculated := derive(c.Base)
	for _, key := range ColorKeys() {
		actual := GetColor(colors, key)
		if actual != "" && !strings.EqualFold(string(actual), string(GetColor(calculated, key))) {
			if c.Overrides == nil {
				c.Overrides = map[string]ui.Color{}
			}

			c.Overrides[key] = actual
		}
	}

	return c
}

var colorsType = reflect.TypeFor[ui.Colors]()

// ColorKeys returns the json names of all fields of [ui.Colors] in their declared order.
func ColorKeys() []string {
	keys := make([]string, 0, colorsType.NumField())
	for i := range colorsType.NumField() {
		keys = append(keys, jsonName(colorsType.Field(i)))
	}

	return keys
}

// GetColor returns the color of the field with the given json name or the empty color, if there is none.
func GetColor(colors ui.Colors, key string) ui.Color {
	idx, ok := colorFieldIndex(key)
	if !ok {
		return ""
	}

	return reflect.ValueOf(colors).Field(idx).Interface().(ui.Color)
}

// SetColor sets the color of the field with the given json name and reports false, if there is none.
func SetColor(colors *ui.Colors, key string, color ui.Color) bool {
	idx, ok := colorFieldIndex(key)
	if !ok {
		return false
	}

	reflect.ValueOf(colors).Elem().Field(idx).Set(reflect.ValueOf(color))
	return true
}

func colorFieldIndex(key string) (int, bool) {
	for i := range colorsType.NumField() {
		if jsonName(colorsType.Field(i)) == key {
			return i, true
		}
	}

	return 0, false
}

func jsonName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}

	return name
}
