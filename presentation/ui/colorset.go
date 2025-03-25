package ui

import (
	"go.wdy.de/nago/presentation/core"
)

// Colors defines a themes color set. See also https://wiki.worldiety.net/books/design-system-ora/page/farbsystem.
type Colors struct {

	// M0 defines the primary source color and is used for a harmonic overall color expression.
	M0 Color `json:"M0"`
	// M1 is usually used for the background.
	M1 Color `json:"M1"`
	// M2 is usually used for the background of a first level container.
	M2 Color `json:"M2"`
	// M3 is usually used for a card bottom.
	M3 Color `json:"M3"`
	// M4 is usually used for a card body.
	M4 Color `json:"M4"`
	// M5 may be used as a line or dot color.
	M5 Color `json:"M5"`
	// M6 is for hovered containers.
	M6 Color `json:"M6"`
	// M7 is for Text or muted icons.
	M7 Color `json:"M7"`
	// M8 is for normal Text or Icons.
	M8 Color `json:"M8"`
	// M9 is for card tops.
	M9 Color `json:"M9"`

	// A0 is the secondary source color and represents non-area color accents.
	A0 Color `json:"A0"`
	// A1 is for progressbars, charts, some accented headlines or borders.
	A1 Color `json:"A1"`
	// A2 is still tbd.
	A2 Color `json:"A2"`

	// I0 is the tertiary source color and represents a non-area interactive element color accent.
	I0 Color `json:"I0"`
	// I1 is used for buttons.
	I1 Color `json:"I1"`

	// (E)rror describes a negative or a destructive (S)emantic intention. In Western Europe usually red. Use it, when the
	// user cannot continue normally and has to fix the problem first.
	Error Color `json:"SE0"`

	// (W)arning describes a critical condition. In Western Europe usually yellow. Use it to warn on situations,
	// which may result in a future error condition.
	Warning Color `json:"SW0"`

	// (G)ood describes a positive condition or a confirming intention. In Western Europe usually green.
	// Use it to symbolize something which has been successfully applied.
	Good Color `json:"SG0"`

	// Informati(V)e shall be used to highlight something, which just changed. E.g. a newly added component or
	// a recommendation from a system. Do not use it to highlight text. In Western Europe usually blue.
	Informative Color `json:"SV0"`

	// Disabled defines an otherwise interactive colored area color but disabled (I)nput.
	Disabled Color `json:"SI0"`

	// DisabledText defines the color of (T)ext which has been rendered on a disabled color.
	DisabledText Color `json:"ST0"`
}

func (c Colors) Valid() bool {
	return c.M0 != "" && c.M1 != "" && c.M2 != "" && c.M3 != "" && c.M4 != "" &&
		c.M5 != "" && c.M6 != "" && c.M7 != "" && c.M8 != "" && c.M9 != "" &&
		c.A0 != "" && c.A1 != "" && c.A2 != "" &&
		c.I0 != "" && c.I1 != "" &&
		c.Error != "" && c.Warning != "" && c.Good != "" &&
		c.Informative != "" && c.Disabled != "" && c.DisabledText != ""
}

func (c Colors) Default(scheme core.ColorScheme) core.ColorSet {
	// this is done correctly in application/theme/uc_read_colors
	var m, a, i Color
	m, a, i = "#8462DA", "#A6A5C2", "#EF8A97"
	return Colors{
		M0:           m,
		M1:           m,
		M2:           m,
		M3:           m,
		M4:           m,
		M5:           m,
		M6:           m,
		M7:           m,
		M8:           m,
		M9:           m,
		A0:           a,
		A1:           a,
		A2:           a,
		I0:           i,
		I1:           i,
		Error:        "",
		Warning:      "",
		Good:         "",
		Informative:  "",
		Disabled:     "",
		DisabledText: "",
	}
}

func (c Colors) Namespace() core.NamespaceName {
	return "ora"
}
