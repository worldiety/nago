// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import (
	"math"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/pkg/xcolor"
	"golang.org/x/image/colornames"
	"golang.org/x/text/language"
)

type namedColor struct {
	name string // identifier as used by golang.org/x/image/colornames
	str  i18n.StrHnd
	rgb  xcolor.RGBA8
	lab  xcolor.LAB
}

func newNamedColor(name, en, de string) namedColor {
	return namedColor{
		name: name,
		str: i18n.MustString(
			i18n.Key("nago.colorpicker.color."+name),
			i18n.Values{
				language.English: en,
				language.German:  de,
			},
			i18n.LocalizationHint("Name of a color. It is used to describe a color to the user, e.g. for accessibility labels."),
		),
	}
}

// namedColors contains all colors of golang.org/x/image/colornames in alphabetical order. If multiple names
// share the same color value (e.g. aqua and cyan), the first one wins on lookup.
var namedColors = initNamedColors([]namedColor{
	newNamedColor("aliceblue", "Alice Blue", "Eisblau"),
	newNamedColor("antiquewhite", "Antique White", "Antikweiß"),
	newNamedColor("aqua", "Aqua", "Aqua"),
	newNamedColor("aquamarine", "Aquamarine", "Aquamarin"),
	newNamedColor("azure", "Azure", "Himmelblau"),
	newNamedColor("beige", "Beige", "Beige"),
	newNamedColor("bisque", "Bisque", "Biskuit"),
	newNamedColor("black", "Black", "Schwarz"),
	newNamedColor("blanchedalmond", "Blanched Almond", "Mandelweiß"),
	newNamedColor("blue", "Blue", "Blau"),
	newNamedColor("blueviolet", "Blue Violet", "Blauviolett"),
	newNamedColor("brown", "Brown", "Braun"),
	newNamedColor("burlywood", "Burlywood", "Gelbbraun"),
	newNamedColor("cadetblue", "Cadet Blue", "Kadettenblau"),
	newNamedColor("chartreuse", "Chartreuse", "Hellgrün"),
	newNamedColor("chocolate", "Chocolate", "Schokolade"),
	newNamedColor("coral", "Coral", "Koralle"),
	newNamedColor("cornflowerblue", "Cornflower Blue", "Kornblumenblau"),
	newNamedColor("cornsilk", "Cornsilk", "Mais"),
	newNamedColor("crimson", "Crimson", "Karmesinrot"),
	newNamedColor("cyan", "Cyan", "Cyan"),
	newNamedColor("darkblue", "Dark Blue", "Dunkelblau"),
	newNamedColor("darkcyan", "Dark Cyan", "Dunkelcyan"),
	newNamedColor("darkgoldenrod", "Dark Goldenrod", "Dunkle Goldrute"),
	newNamedColor("darkgray", "Dark Gray", "Dunkelgrau"),
	newNamedColor("darkgreen", "Dark Green", "Dunkelgrün"),
	newNamedColor("darkgrey", "Dark Grey", "Dunkelgrau"),
	newNamedColor("darkkhaki", "Dark Khaki", "Dunkelkhaki"),
	newNamedColor("darkmagenta", "Dark Magenta", "Dunkelmagenta"),
	newNamedColor("darkolivegreen", "Dark Olive Green", "Dunkles Olivgrün"),
	newNamedColor("darkorange", "Dark Orange", "Dunkelorange"),
	newNamedColor("darkorchid", "Dark Orchid", "Dunkle Orchidee"),
	newNamedColor("darkred", "Dark Red", "Dunkelrot"),
	newNamedColor("darksalmon", "Dark Salmon", "Dunkles Lachsrosa"),
	newNamedColor("darkseagreen", "Dark Sea Green", "Dunkles Seegrün"),
	newNamedColor("darkslateblue", "Dark Slate Blue", "Dunkles Schieferblau"),
	newNamedColor("darkslategray", "Dark Slate Gray", "Dunkles Schiefergrau"),
	newNamedColor("darkslategrey", "Dark Slate Grey", "Dunkles Schiefergrau"),
	newNamedColor("darkturquoise", "Dark Turquoise", "Dunkeltürkis"),
	newNamedColor("darkviolet", "Dark Violet", "Dunkelviolett"),
	newNamedColor("deeppink", "Deep Pink", "Tiefrosa"),
	newNamedColor("deepskyblue", "Deep Sky Blue", "Tiefes Himmelblau"),
	newNamedColor("dimgray", "Dim Gray", "Mattgrau"),
	newNamedColor("dimgrey", "Dim Grey", "Mattgrau"),
	newNamedColor("dodgerblue", "Dodger Blue", "Dodger-Blau"),
	newNamedColor("firebrick", "Firebrick", "Ziegelrot"),
	newNamedColor("floralwhite", "Floral White", "Blütenweiß"),
	newNamedColor("forestgreen", "Forest Green", "Waldgrün"),
	newNamedColor("fuchsia", "Fuchsia", "Fuchsia"),
	newNamedColor("gainsboro", "Gainsboro", "Gainsboro"),
	newNamedColor("ghostwhite", "Ghost White", "Geisterweiß"),
	newNamedColor("gold", "Gold", "Gold"),
	newNamedColor("goldenrod", "Goldenrod", "Goldrute"),
	newNamedColor("gray", "Gray", "Grau"),
	newNamedColor("green", "Green", "Grün"),
	newNamedColor("greenyellow", "Green Yellow", "Grüngelb"),
	newNamedColor("grey", "Grey", "Grau"),
	newNamedColor("honeydew", "Honeydew", "Honigmelone"),
	newNamedColor("hotpink", "Hot Pink", "Leuchtendes Rosa"),
	newNamedColor("indianred", "Indian Red", "Indischrot"),
	newNamedColor("indigo", "Indigo", "Indigo"),
	newNamedColor("ivory", "Ivory", "Elfenbein"),
	newNamedColor("khaki", "Khaki", "Khaki"),
	newNamedColor("lavender", "Lavender", "Lavendel"),
	newNamedColor("lavenderblush", "Lavender Blush", "Lavendelrosa"),
	newNamedColor("lawngreen", "Lawn Green", "Rasengrün"),
	newNamedColor("lemonchiffon", "Lemon Chiffon", "Zitronengelb"),
	newNamedColor("lightblue", "Light Blue", "Hellblau"),
	newNamedColor("lightcoral", "Light Coral", "Helles Korallenrot"),
	newNamedColor("lightcyan", "Light Cyan", "Hellcyan"),
	newNamedColor("lightgoldenrodyellow", "Light Goldenrod Yellow", "Helles Goldrutengelb"),
	newNamedColor("lightgray", "Light Gray", "Hellgrau"),
	newNamedColor("lightgreen", "Light Green", "Hellgrün"),
	newNamedColor("lightgrey", "Light Grey", "Hellgrau"),
	newNamedColor("lightpink", "Light Pink", "Hellrosa"),
	newNamedColor("lightsalmon", "Light Salmon", "Helles Lachsrosa"),
	newNamedColor("lightseagreen", "Light Sea Green", "Helles Seegrün"),
	newNamedColor("lightskyblue", "Light Sky Blue", "Helles Himmelblau"),
	newNamedColor("lightslategray", "Light Slate Gray", "Helles Schiefergrau"),
	newNamedColor("lightslategrey", "Light Slate Grey", "Helles Schiefergrau"),
	newNamedColor("lightsteelblue", "Light Steel Blue", "Helles Stahlblau"),
	newNamedColor("lightyellow", "Light Yellow", "Hellgelb"),
	newNamedColor("lime", "Lime", "Limette"),
	newNamedColor("limegreen", "Lime Green", "Limettengrün"),
	newNamedColor("linen", "Linen", "Leinen"),
	newNamedColor("magenta", "Magenta", "Magenta"),
	newNamedColor("maroon", "Maroon", "Kastanienbraun"),
	newNamedColor("mediumaquamarine", "Medium Aquamarine", "Mittleres Aquamarin"),
	newNamedColor("mediumblue", "Medium Blue", "Mittelblau"),
	newNamedColor("mediumorchid", "Medium Orchid", "Mittlere Orchidee"),
	newNamedColor("mediumpurple", "Medium Purple", "Mittleres Lila"),
	newNamedColor("mediumseagreen", "Medium Sea Green", "Mittleres Seegrün"),
	newNamedColor("mediumslateblue", "Medium Slate Blue", "Mittleres Schieferblau"),
	newNamedColor("mediumspringgreen", "Medium Spring Green", "Mittleres Frühlingsgrün"),
	newNamedColor("mediumturquoise", "Medium Turquoise", "Mitteltürkis"),
	newNamedColor("mediumvioletred", "Medium Violet Red", "Mittleres Violettrot"),
	newNamedColor("midnightblue", "Midnight Blue", "Mitternachtsblau"),
	newNamedColor("mintcream", "Mint Cream", "Minzcreme"),
	newNamedColor("mistyrose", "Misty Rose", "Altrosa"),
	newNamedColor("moccasin", "Moccasin", "Mokassin"),
	newNamedColor("navajowhite", "Navajo White", "Navajoweiß"),
	newNamedColor("navy", "Navy", "Marineblau"),
	newNamedColor("oldlace", "Old Lace", "Alte Spitze"),
	newNamedColor("olive", "Olive", "Oliv"),
	newNamedColor("olivedrab", "Olive Drab", "Olivgrau"),
	newNamedColor("orange", "Orange", "Orange"),
	newNamedColor("orangered", "Orange Red", "Orangerot"),
	newNamedColor("orchid", "Orchid", "Orchidee"),
	newNamedColor("palegoldenrod", "Pale Goldenrod", "Blasse Goldrute"),
	newNamedColor("palegreen", "Pale Green", "Blassgrün"),
	newNamedColor("paleturquoise", "Pale Turquoise", "Blasstürkis"),
	newNamedColor("palevioletred", "Pale Violet Red", "Blasses Violettrot"),
	newNamedColor("papayawhip", "Papaya Whip", "Papayacreme"),
	newNamedColor("peachpuff", "Peach Puff", "Pfirsich"),
	newNamedColor("peru", "Peru", "Peru"),
	newNamedColor("pink", "Pink", "Rosa"),
	newNamedColor("plum", "Plum", "Pflaume"),
	newNamedColor("powderblue", "Powder Blue", "Puderblau"),
	newNamedColor("purple", "Purple", "Lila"),
	newNamedColor("red", "Red", "Rot"),
	newNamedColor("rosybrown", "Rosy Brown", "Rosiges Braun"),
	newNamedColor("royalblue", "Royal Blue", "Königsblau"),
	newNamedColor("saddlebrown", "Saddle Brown", "Sattelbraun"),
	newNamedColor("salmon", "Salmon", "Lachsrosa"),
	newNamedColor("sandybrown", "Sandy Brown", "Sandbraun"),
	newNamedColor("seagreen", "Sea Green", "Seegrün"),
	newNamedColor("seashell", "Seashell", "Muschelweiß"),
	newNamedColor("sienna", "Sienna", "Siena"),
	newNamedColor("silver", "Silver", "Silber"),
	newNamedColor("skyblue", "Sky Blue", "Himmelblau"),
	newNamedColor("slateblue", "Slate Blue", "Schieferblau"),
	newNamedColor("slategray", "Slate Gray", "Schiefergrau"),
	newNamedColor("slategrey", "Slate Grey", "Schiefergrau"),
	newNamedColor("snow", "Snow", "Schneeweiß"),
	newNamedColor("springgreen", "Spring Green", "Frühlingsgrün"),
	newNamedColor("steelblue", "Steel Blue", "Stahlblau"),
	newNamedColor("tan", "Tan", "Hautfarben"),
	newNamedColor("teal", "Teal", "Petrol"),
	newNamedColor("thistle", "Thistle", "Distel"),
	newNamedColor("tomato", "Tomato", "Tomatenrot"),
	newNamedColor("turquoise", "Turquoise", "Türkis"),
	newNamedColor("violet", "Violet", "Violett"),
	newNamedColor("wheat", "Wheat", "Weizen"),
	newNamedColor("white", "White", "Weiß"),
	newNamedColor("whitesmoke", "White Smoke", "Rauchweiß"),
	newNamedColor("yellow", "Yellow", "Gelb"),
	newNamedColor("yellowgreen", "Yellow Green", "Gelbgrün"),
})

func initNamedColors(colors []namedColor) []namedColor {
	for i := range colors {
		c, ok := colornames.Map[colors[i].name]
		if !ok {
			panic("colorpicker: unknown color name: " + colors[i].name)
		}

		colors[i].rgb = xcolor.RGBA8{c.R, c.G, c.B, 255}
		colors[i].lab = colors[i].rgb.RGBA().LAB()
	}

	return colors
}

var (
	strColorNameWithAlphaX = i18n.MustVarString(
		"nago.colorpicker.color_name_with_alpha_x",
		i18n.Values{
			language.English: "{name} ({alpha}%)",
			language.German:  "{name} ({alpha} %)",
		},
		i18n.LocalizationHint("Name of a color followed by its opacity in percent, e.g. 'Red (50%)'."),
	)
)

// ColorName maps the given hex color code (e.g. #FF0000 or #FF000080) to a localized color name, based on the
// names defined by golang.org/x/image/colornames. If the color does not exactly match a named color, the
// perceptually nearest named color (CIELAB distance) is returned. If the opacity is below 100%, it is appended
// in percent, e.g. "Red (50%)". If the hex code is invalid (might be rgba or a variable name), it is returned as is.
func ColorName(bnd i18n.Bundler, hex string) string {
	c, alpha, ok := lookupNamedColor(hex)
	if !ok {
		return hex
	}

	name := c.str.Get(bnd)
	if alpha == 255 {
		return name
	}

	return strColorNameWithAlphaX.Get(bnd, i18n.String("name", name), i18n.Int("alpha", alphaPercent(alpha)))
}

// alphaPercent converts the alpha channel into percent. A value below 255 never results in 100.
func alphaPercent(alpha uint8) int {
	p := int(math.Round(float64(alpha) * 100 / 255))
	if alpha < 255 && p == 100 {
		p = 99
	}

	return p
}

// ColorNameKey is like [ColorName] but returns the non-localized identifier as used by
// golang.org/x/image/colornames (e.g. "red" or "cornflowerblue"). The alpha channel is ignored.
// Returns false, if the hex code is invalid.
func ColorNameKey(hex string) (string, bool) {
	c, _, ok := lookupNamedColor(hex)
	if !ok {
		return "", false
	}

	return c.name, true
}

func lookupNamedColor(hex string) (namedColor, uint8, bool) {
	rgba, err := xcolor.ParseHex(strings.TrimSpace(hex))
	if err != nil {
		return namedColor{}, 0, false
	}

	alpha := rgba[3]
	rgba[3] = 255
	for _, c := range namedColors {
		if c.rgb == rgba {
			return c, alpha, true
		}
	}

	lab := rgba.RGBA().LAB()
	best := namedColors[0]
	bestDist := math.MaxFloat64
	for _, c := range namedColors {
		dl := float64(lab[0] - c.lab[0])
		da := float64(lab[1] - c.lab[1])
		db := float64(lab[2] - c.lab[2])
		if d := dl*dl + da*da + db*db; d < bestDist {
			bestDist = d
			best = c
		}
	}

	return best, alpha, true
}
