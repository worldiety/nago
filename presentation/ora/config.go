package ora

import (
	"math"
	"strconv"
	"strings"
)

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ColorScheme string

const (
	LightMode ColorScheme = "light"
	DarkMode  ColorScheme = "dark"
)

// ConfigurationRequested is issued by the frontend to get the applications general configuration.
// A backend developer has potentially defined a lot of configuration details about the application.
// For example, there may be a color theme, customized icons, image resources, an application name and the available set of navigations, launch intents or other meta information.
// It is expected, that this only happens once during initialization of the frontend process.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ConfigurationRequested struct {
	Type           EventType   `json:"type" value:"ConfigurationRequested"`
	AcceptLanguage string      `json:"acceptLanguage"`
	ColorScheme    ColorScheme `json:"colorScheme" description:"Color scheme hint which the frontend has picked. This may reduce graphical glitches, if the backend creates images or webview resources for the frontend."`
	WindowInfo     WindowInfo  `json:"windowInfo"`
	RequestId      RequestId   `json:"r" `
	event
}

func (e ConfigurationRequested) ReqID() RequestId {
	return e.RequestId
}

// A ConfigurationDefined event is the response to a [ConfigurationRequested] event.
// According to the locale request, string and svg resources can be localized by the backend.
// The returned locale is the actually picked locale from the requested locale query string.
//
// It looks quite obfuscated, however this minified version is intentional, because it may succeed each transaction call.
// A frontend may request acknowledges for each event, e.g. while typing in a text field, so this premature optimization is likely a win.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ConfigurationDefined struct {
	Type               EventType `json:"type" value:"ConfigurationDefined"`
	ApplicationID      string    `json:"applicationID"`
	ApplicationName    string    `json:"applicationName"`
	ApplicationVersion string    `json:"applicationVersion"`
	AvailableLocales   []string  `json:"availableLocales"`
	ActiveLocale       string    `json:"activeLocale"`
	Themes             Themes    `json:"themes"`
	Resources          Resources `json:"resources"`
	RequestId          RequestId `json:"r"`
	event
}

func (e ConfigurationDefined) ReqID() RequestId {
	return e.RequestId
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Themes struct {
	Dark         Theme `json:"dark"`
	Light        Theme `json:"light"`
	HighContrast Theme `json:"highContrast"`
	Protanopie   Theme `json:"protanopie"`
	Deuteranopie Theme `json:"deuteranopie"`
	Tritanopie   Theme `json:"tritanopie"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Theme struct {
	Colors Colors `json:"colors"`
}

type ThemeOption interface {
	apply(*Theme)
}

type themeFunc func(*Theme)

func (f themeFunc) apply(dst *Theme) {
	f(dst)
}

func PrimaryColor(color Color) ThemeOption {
	return themeFunc(func(theme *Theme) {
		theme.Colors.Primary = HSL(color.H, color.S, color.L)
		theme.Colors.PrimaryTen = HSL(color.H, color.S, 10)
		theme.Colors.PrimaryTwelve = HSL(color.H, color.S, 12)
		theme.Colors.PrimaryFourteen = HSL(color.H, color.S, 14)
		theme.Colors.PrimarySeventeen = HSL(color.H, color.S, 17)
		theme.Colors.PrimaryTwentyTwo = HSL(color.H, color.S, 22)
		theme.Colors.PrimaryThirty = HSL(color.H, color.S, 30)
		theme.Colors.PrimarySixty = HSL(color.H, color.S, 60)
		theme.Colors.PrimarySeventy = HSL(color.H, color.S, 70)
		theme.Colors.PrimaryEightyThree = HSL(color.H, color.S, 83)
		theme.Colors.PrimaryEightySeven = HSL(color.H, color.S, 87)
		theme.Colors.PrimaryNinety = HSL(color.H, color.S, 90)
		theme.Colors.PrimaryNinetyTwo = HSL(color.H, color.S, 92)
		theme.Colors.PrimaryNinetyFour = HSL(color.H, color.S, 94)
		theme.Colors.PrimaryNinetySix = HSL(color.H, color.S, 96)
		theme.Colors.PrimaryNinetyEight = HSL(color.H, color.S, 98)
	})
}

func SecondaryColor(color Color) ThemeOption {
	return themeFunc(func(theme *Theme) {
		theme.Colors.Secondary = color
	})
}

func TertiaryColor(color Color) ThemeOption {
	return themeFunc(func(theme *Theme) {
		theme.Colors.Tertiary = color
	})
}

func GenerateTheme(opts ...ThemeOption) Theme {
	var theme Theme
	for _, opt := range opts {
		opt.apply(&theme)
	}

	return theme
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Colors struct {
	Primary            Color `json:"primary"`
	PrimaryTen         Color `json:"primary10"`
	PrimaryTwelve      Color `json:"primary12"`
	PrimaryFourteen    Color `json:"primary14"`
	PrimarySeventeen   Color `json:"primary17"`
	PrimaryTwentyTwo   Color `json:"primary22"`
	PrimaryThirty      Color `json:"primary30"`
	PrimarySixty       Color `json:"primary60"`
	PrimarySeventy     Color `json:"primary70"`
	PrimaryEightyThree Color `json:"primary83"`
	PrimaryEightySeven Color `json:"primary87"`
	PrimaryNinety      Color `json:"primary90"`
	PrimaryNinetyTwo   Color `json:"primary92"`
	PrimaryNinetyFour  Color `json:"primary94"`
	PrimaryNinetySix   Color `json:"primary96"`
	PrimaryNinetyEight Color `json:"primary98"`
	Secondary          Color `json:"secondary"`
	Tertiary           Color `json:"tertiary"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Color struct { // TODO this should be a HSL struct and may be introduce a Color interface and the ranges do not look idiomatic, probably must be 0-1 for each component
	H float64 `json:"h"` // degree from 0 - 360
	S float64 `json:"s"` // percent from 0 to 100
	L float64 `json:"l"` // percent from 0 to 100
}

func HSL(hueAngle float64, saturationPercentage float64, lightnessPercentage float64) Color {
	h := math.Max(0, math.Min(hueAngle, 360))
	s := math.Max(0, math.Min(saturationPercentage, 100))
	l := math.Max(0, math.Min(lightnessPercentage, 100))
	return Color{h, s, l}
}

func MustParseHSL(hex string) Color {
	if strings.HasPrefix(hex, "#") {
		hex = hex[1:]
	}
	r, g, b, _ := hexToRGBA(hex)
	h, s, l := rgbToHSL(r, g, b)
	return Color{H: h * 360, S: s * 100, L: l * 100} // convert from conventional to our "human" format
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Resources struct {
	SVG map[RIDSVG]SVG `json:"svgs"`
}

func hexToRGBA(hex string) (r, g, b, a uint8) {
	if len(hex) == 6 {
		hex += "FF"
	}
	rgba, _ := strconv.ParseUint(hex, 16, 32)
	r = uint8((rgba >> 24) & 0xFF)
	g = uint8((rgba >> 16) & 0xFF)
	b = uint8((rgba >> 8) & 0xFF)
	a = uint8(rgba & 0xFF)
	return
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	min := min(min(rf, gf), bf)
	max := max(max(rf, gf), bf)

	l = (max + min) / 2

	if max == min {
		h, s = 0.0, 0.0
	} else {
		d := max - min

		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if g < b {
				h += 6.0
			}
		case gf:
			h = 2.0 + (bf-rf)/d
		case bf:
			h = 4.0 + (rf-gf)/d
		}
		h /= 6.0
	}

	return
}
