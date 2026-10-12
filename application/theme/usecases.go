// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type LogoType int

const (
	// NavigationBarIcon allows different formats and is limited in its height first.
	// It should not have a transparent border, to avoid visual padding issues in the final layout.
	NavigationBarIcon LogoType = iota + 1

	// AppIcon must be quadratic. Recommended resolution is 512x512 or 1024x1024. It should not have a transparent
	// border, to avoid visual padding issues in the final layout. It may be used as a favicon or shown
	// during the registration process.
	AppIcon
)

type UpdateDefaultLogo func(subject auth.Subject, logoType LogoType, scheme core.ColorScheme, img []byte) error

type BaseColors struct {
	Main        ui.Color
	Interactive ui.Color
	Accent      ui.Color
}
type DeriveColors func(colors BaseColors) ui.Colors

// Calculations provides implementations for different color derivation functions which usually accept BaseColors.
type Calculations struct {
	DarkMode      DeriveColors
	TrueDarkMode  DeriveColors
	LightMode     DeriveColors
	TrueLightMode DeriveColors
}

// HasColors returns true, if [Settings.Colors] Dark and Light [ui.Colors] are both valid, i.e. if colors have been
// stored. Otherwise, returns false. To propose the colors of an application, do not store them, if HasColors returns
// false, but use [go.wdy.de/nago/application.Configurator.SetThemeDefaults] instead.
type HasColors func(subject auth.Subject) (bool, error)

// ResetColors removes the stored colors and the [Customization], so that the proposal of the application applies
// again, see [Defaults]. HasColors will return false.
type ResetColors func(subject auth.Subject) error

// UpdateColors stores the given colors, which win over the proposal of the application. It discards the
// [Customization]. Use [UpdateCustomization] to store the choice of a user.
type UpdateColors func(subject auth.Subject, colors Colors) error

// ReadColors returns the stored colors or the proposal of the application, see [Defaults].
type ReadColors func(subject auth.Subject) (Colors, error)

// ReadCustomization returns the stored choice of the user. If there is none, it is derived from the colors, which
// [ReadColors] returns, see [CustomizationOf].
type ReadCustomization func(subject auth.Subject) (Customization, error)

// UpdateCustomization stores the choice of the user and the colors calculated from it, which win over the proposal
// of the application.
type UpdateCustomization func(subject auth.Subject, customization Customization) error

type UpdateFonts func(subject auth.Subject, fonts core.Fonts) error
type ReadFonts func(subject auth.Subject) (core.Fonts, error)

type UseCases struct {
	Calculations Calculations
	UpdateColors UpdateColors
	ReadColors   ReadColors
	HasColors    HasColors
	ResetColors  ResetColors
	UpdateFonts  UpdateFonts
	ReadFonts    ReadFonts

	ReadCustomization   ReadCustomization
	UpdateCustomization UpdateCustomization
}

// NewUseCases creates the use cases. The defaults may be nil and are read, whenever no colors have been stored.
func NewUseCases(bus events.Bus, loadGlobal settings.LoadGlobal, storeGlobal settings.StoreGlobal, defaults Defaults) UseCases {
	return UseCases{
		Calculations: Calculations{
			DarkMode:      DarkMode,
			TrueDarkMode:  TrueDarkMode,
			LightMode:     LightMode,
			TrueLightMode: TrueLightMode,
		},
		UpdateColors: NewUpdateColors(bus, loadGlobal, storeGlobal),
		ReadColors:   NewReadColors(loadGlobal, defaults),
		HasColors:    NewHasColors(loadGlobal),
		ResetColors:  NewResetColors(bus, loadGlobal, storeGlobal),
		UpdateFonts:  NewUpdateFonts(bus, loadGlobal, storeGlobal),
		ReadFonts:    NewReadFonts(loadGlobal),

		ReadCustomization:   NewReadCustomization(loadGlobal, defaults),
		UpdateCustomization: NewUpdateCustomization(bus, loadGlobal, storeGlobal),
	}
}
