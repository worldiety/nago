package ui

import "go.wdy.de/nago/presentation/core"

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

func DefaultColors(scheme core.ColorScheme, main, accent, interactive Color) Colors {
	switch scheme {
	case core.Dark:
		return darkColors(main, accent, interactive)
	default:
		return lightColors(main, accent, interactive)
	}
}

func darkColors(main, accent, interactive Color) Colors {
	var c Colors
	m := mustParseHSL(string(main))
	c.M0 = main
	c.M1 = m.Brightness(10).RGBHex()
	c.M2 = m.Brightness(12).RGBHex()
	c.M3 = m.Brightness(22).RGBHex()
	c.M4 = m.Brightness(17).RGBHex()
	c.M5 = m.Brightness(30).RGBHex()
	c.M6 = m.Brightness(22).RGBHex()
	c.M7 = m.Brightness(60).RGBHex()
	c.M8 = m.Brightness(90).RGBHex()
	c.M9 = m.Brightness(14).RGBHex()

	_ = mustParseHSL(string(accent))
	c.A0 = accent
	c.A1 = accent
	c.A2 = accent.WithBrightness(80)

	_ = mustParseHSL(string(interactive))
	c.I0 = interactive
	c.I1 = interactive

	c.Error = "#F47954"
	c.Warning = "#F7A823"
	c.Informative = "#17428C"
	c.Good = "#2BCA73"
	c.Disabled = "#E2E2E2"
	c.DisabledText = "#848484"

	return c
}

func lightColors(main, accent, interactive Color) Colors {
	var c Colors
	m := mustParseHSL(string(main))
	c.M0 = main
	c.M1 = m.Brightness(98).RGBHex()
	c.M2 = m.Brightness(96).RGBHex()
	c.M3 = m.Brightness(90).RGBHex()
	c.M4 = m.Brightness(94).RGBHex()
	c.M5 = m.Brightness(70).RGBHex()
	c.M6 = m.Brightness(90).RGBHex()
	c.M7 = m.Brightness(60).RGBHex()
	c.M8 = m.Brightness(10).RGBHex()
	c.M9 = m.Brightness(92).RGBHex()

	_ = mustParseHSL(string(accent))
	c.A0 = accent
	c.A1 = accent
	c.A2 = accent.WithBrightness(80)

	_ = mustParseHSL(string(interactive))
	c.I0 = interactive
	c.I1 = interactive

	c.Error = "#F47954"
	c.Warning = "#F7A823"
	c.Informative = "#17428C"
	c.Good = "#2BCA73"
	c.Disabled = "#E2E2E2"
	c.DisabledText = "#848484"

	return c
}

func (c Colors) Default(scheme core.ColorScheme) core.ColorSet {
	var m, a, i Color
	m, a, i = "#8462DA", "#A6A5C2", "#EF8A97"
	return DefaultColors(scheme, m, a, i)
}

func (c Colors) Namespace() core.NamespaceName {
	return "ora"
}
