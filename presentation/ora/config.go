package ora

import (
	"encoding/json"
	"fmt"
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

func DefaultTheme() Theme {
	return Theme{
		Colors: Colors{
			Black:         RGB(0x000000),
			White:         RGB(0xFFFFFF),
			Primary:       RGB(0x1B8C30),
			Interactive:   RGB(0xF7A823),
			AlertNegative: RGB(0xFF543E),
			AlertPositive: RGB(0x54FF3E),
		},
	}
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Colors struct {
	Black         Color `json:"black"`
	White         Color `json:"white"`
	Primary       Color `json:"primary"`
	Interactive   Color `json:"interactive"`
	AlertNegative Color `json:"alertNegative"`
	AlertPositive Color `json:"alertPositive"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

func (c Color) MarshalJSON() ([]byte, error) {
	if c.A == 0xFF {
		return json.Marshal(fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B))
	}

	return json.Marshal(fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A))
}

func RGBA(color uint32) Color {
	r := uint8((color >> 24) & 0xFF)
	g := uint8((color >> 16) & 0xFF)
	b := uint8((color >> 8) & 0xFF)
	a := uint8(color & 0xFF)
	return Color{r, g, b, a}
}

func RGB(color uint32) Color {
	r := uint8((color >> 16) & 0xFF)
	g := uint8((color >> 8) & 0xFF)
	b := uint8((color >> 0) & 0xFF)
	a := uint8(0xFF)
	return Color{r, g, b, a}
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Resources struct {
	SVG map[RIDSVG]SVG `json:"svgs"`
}
