package ora

import (
	"math"
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

func LightTheme() Theme {
	primaryHueAngle := uint16(38)
	primarySaturationPercentage := uint8(93)
	primaryLightnessPercentage := uint8(55)

	return generateTheme(
		primaryHueAngle,
		primarySaturationPercentage,
		primaryLightnessPercentage,
		120,
		100,
		50,
		240,
		100,
		50,
	)
}

func DarkTheme() Theme {
	primaryHueAngle := uint16(131)
	primarySaturationPercentage := uint8(68)
	primaryLightnessPercentage := uint8(33)

	return generateTheme(
		primaryHueAngle,
		primarySaturationPercentage,
		primaryLightnessPercentage,
		120,
		100,
		50,
		240,
		100,
		50,
	)
}

func generateTheme(
	primaryHueAngle uint16,
	primarySaturationPercentage uint8,
	primaryLightnessPercentage uint8,
	secondaryHueAngle uint16,
	secondarySaturationPercentage uint8,
	secondaryLightnessPercentage uint8,
	tertiaryHueAngle uint16,
	tertiarySaturationPercentage uint8,
	tertiaryLightnessPercentage uint8,
) Theme {
	return Theme{
		Colors: Colors{
			Primary:            HSL(primaryHueAngle, primarySaturationPercentage, primaryLightnessPercentage),
			PrimaryTen:         HSL(primaryHueAngle, primarySaturationPercentage, 10),
			PrimaryTwelve:      HSL(primaryHueAngle, primarySaturationPercentage, 12),
			PrimaryFourteen:    HSL(primaryHueAngle, primarySaturationPercentage, 14),
			PrimarySeventeen:   HSL(primaryHueAngle, primarySaturationPercentage, 17),
			PrimaryTwentyTwo:   HSL(primaryHueAngle, primarySaturationPercentage, 22),
			PrimaryThirty:      HSL(primaryHueAngle, primarySaturationPercentage, 30),
			PrimarySixty:       HSL(primaryHueAngle, primarySaturationPercentage, 60),
			PrimarySeventy:     HSL(primaryHueAngle, primarySaturationPercentage, 70),
			PrimaryEightyThree: HSL(primaryHueAngle, primarySaturationPercentage, 83),
			PrimaryEightySeven: HSL(primaryHueAngle, primarySaturationPercentage, 87),
			PrimaryNinety:      HSL(primaryHueAngle, primarySaturationPercentage, 90),
			PrimaryNinetyTwo:   HSL(primaryHueAngle, primarySaturationPercentage, 92),
			PrimaryNinetyFour:  HSL(primaryHueAngle, primarySaturationPercentage, 94),
			PrimaryNinetySix:   HSL(primaryHueAngle, primarySaturationPercentage, 96),
			PrimaryNinetyEight: HSL(primaryHueAngle, primarySaturationPercentage, 98),
			Secondary:          HSL(secondaryHueAngle, secondarySaturationPercentage, secondaryLightnessPercentage),
			Tertiary:           HSL(tertiaryHueAngle, tertiarySaturationPercentage, tertiaryLightnessPercentage),
		},
	}
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
type Color struct {
	H uint16 `json:"h"`
	S uint8  `json:"s"`
	L uint8  `json:"l"`
}

func HSL(hueAngle uint16, saturationPercentage uint8, lightnessPercentage uint8) Color {
	h := uint16(math.Min(float64(hueAngle), 360))
	s := uint8(math.Min(float64(saturationPercentage), 100))
	l := uint8(math.Min(float64(lightnessPercentage), 100))
	return Color{h, s, l}
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Resources struct {
	SVG map[RIDSVG]SVG `json:"svgs"`
}
