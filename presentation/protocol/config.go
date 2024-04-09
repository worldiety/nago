package protocol

import (
	"encoding/json"
	"fmt"
)

type ColorScheme string

const (
	LightMode ColorScheme = "light"
	DarkMode  ColorScheme = "dark"
)

type ConfigurationRequested struct {
	Type           EventType   `json:"type" value:"NewConfigurationRequested"`
	AcceptLanguage string      `json:"acceptLanguage"`
	ColorScheme    ColorScheme `json:"colorScheme" description:"Color scheme hint which the frontend has picked. This may reduce graphical glitches, if the backend creates images or webview resources for the frontend."`
	event
}

type ConfigurationDefined struct {
	Type             EventType `json:"type" value:"ConfigurationDefined"`
	ApplicationName  string    `json:"applicationName"`
	AvailableLocales []string  `json:"availableLocales"`
	ActiveLocale     string    `json:"activeLocale"`
	Themes           Themes    `json:"themes"`
	Resources        Resources `json:"resources"`
	event
}

type Themes struct {
	Dark  Theme `json:"dark"`
	Light Theme `json:"light"`
}

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

type Colors struct {
	Black         Color `json:"black"`
	White         Color `json:"white"`
	Primary       Color `json:"primary"`
	Interactive   Color `json:"interactive"`
	AlertNegative Color `json:"alertNegative"`
	AlertPositive Color `json:"alertPositive"`
}

type Color struct {
	R, G, B, A uint8
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

type SVGID int
type SVGSrc string

type Resources struct {
	SVG map[SVGID]SVGSrc `json:"svgs"`
}
