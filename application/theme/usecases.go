package theme

import (
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/ui"
)

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

// HasColors returns true, if [Settings.Colors] Dark and Light [ui.Colors] are both valid. Otherwise, returns false.
// You can check this to initially apply a theme and [Update] the colors or leave it eventually customized by the user.
type HasColors func(subject auth.Subject) (bool, error)

// ResetColors replaces the [Settings.Colors] field with its zero value and writes it into the settings.
// There is no domain event fired, because it is unclear, if the system fallback default or the developers application
// default must be read. Therefore, you have to restart the process, to execute whatever theme logic needs to be applied.
// HasColors will return false.
type ResetColors func(subject auth.Subject) error

type UpdateColors func(subject auth.Subject, colors Colors) error
type ReadColors func(subject auth.Subject) (Colors, error)

type UseCases struct {
	Calculations Calculations
	UpdateColors UpdateColors
	ReadColors   ReadColors
	HasColors    HasColors
	ResetColors  ResetColors
}

func NewUseCases(bus events.Bus, loadGlobal settings.LoadGlobal, storeGlobal settings.StoreGlobal) UseCases {
	return UseCases{
		Calculations: Calculations{
			DarkMode:      DarkMode,
			TrueDarkMode:  TrueDarkMode,
			LightMode:     LightMode,
			TrueLightMode: TrueLightMode,
		},
		UpdateColors: NewUpdateColors(bus, loadGlobal, storeGlobal),
		ReadColors:   NewReadColors(loadGlobal),
		HasColors:    NewHasColors(loadGlobal),
		ResetColors:  NewResetColors(bus, loadGlobal, storeGlobal),
	}
}
