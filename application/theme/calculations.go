// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/xcolor"
	"go.wdy.de/nago/presentation/ui"
)

func DarkMode(base BaseColors) ui.Colors {
	c := ui.Colors{
		M9: option.Must(base.Main.WithChromaAndTone(16, 22)),
		M8: option.Must(base.Main.WithChromaAndTone(8, 98)),
		M7: option.Must(base.Main.WithChromaAndTone(8, 75)),
		M6: option.Must(base.Main.WithChromaAndTone(16, 22)),
		M5: option.Must(base.Main.WithChromaAndTone(16, 30)),
		M4: option.Must(base.Main.WithChromaAndTone(16, 20)),
		M3: option.Must(base.Main.WithChromaAndTone(16, 22)),
		M2: option.Must(base.Main.WithChromaAndTone(16, 15)),
		M1: option.Must(base.Main.WithChromaAndTone(16, 5)),
		M0: base.Main,

		I1: option.Must(base.Interactive.WithChromaAndTone(100, 50)),
		I0: base.Interactive,

		A2: option.Must(base.Accent.WithChromaAndTone(8, 75)),
		A1: option.Must(base.Main.WithChromaAndTone(16, 22)),
		A0: base.Accent,

		PrimaryButtonText: calcTextColor(base.Interactive),

		BannerErrorBackground: "#3b1812",
		BannerErrorText:       "#FF543E",
		BannerInfoBackground:  "#1c3b12",
		BannerInfoText:        "#55ff3e",
	}

	withSemanticColors(&c)

	return c
}

func TrueDarkMode(base BaseColors) ui.Colors {
	c := ui.Colors{
		M9: option.Must(base.Main.WithChromaAndTone(8, 14)),
		M8: option.Must(base.Main.WithChromaAndTone(8, 98)),
		M7: option.Must(base.Main.WithChromaAndTone(8, 75)),
		M6: option.Must(base.Main.WithChromaAndTone(8, 22)),
		M5: option.Must(base.Main.WithChromaAndTone(8, 30)),
		M4: option.Must(base.Main.WithChromaAndTone(8, 14)),
		M3: option.Must(base.Main.WithChromaAndTone(8, 17)),
		M2: option.Must(base.Main.WithChromaAndTone(8, 10)),
		M1: option.Must(base.Main.WithChromaAndTone(8, 5)),
		M0: base.Main,

		I1: option.Must(base.Interactive.WithChromaAndTone(16, 22)),
		I0: base.Interactive,

		A2: option.Must(base.Accent.WithChromaAndTone(8, 75)),
		A1: option.Must(base.Main.WithChromaAndTone(16, 22)),
		A0: base.Accent,

		PrimaryButtonText: calcTextColor(base.Interactive),

		BannerErrorBackground: "#3b1812",
		BannerErrorText:       "#FF543E",
		BannerInfoBackground:  "#1c3b12",
		BannerInfoText:        "#55ff3e",
	}

	withSemanticColors(&c)

	return c
}

func LightMode(base BaseColors) ui.Colors {
	c := ui.Colors{
		M9: option.Must(base.Main.WithChromaAndTone(8, 92)),
		M8: option.Must(base.Main.WithChromaAndTone(8, 10)),
		M7: option.Must(base.Main.WithChromaAndTone(8, 60)),
		M6: option.Must(base.Main.WithChromaAndTone(8, 90)),
		M5: option.Must(base.Main.WithChromaAndTone(8, 70)),
		M4: option.Must(base.Main.WithChromaAndTone(8, 94)),
		M3: option.Must(base.Main.WithChromaAndTone(8, 90)),
		M2: option.Must(base.Main.WithChromaAndTone(6, 96)),
		M1: option.Must(base.Main.WithChromaAndTone(4, 98)),
		M0: base.Main,

		I1: option.Must(base.Interactive.WithChromaAndTone(100, 90)),
		I0: base.Interactive,

		A2: option.Must(base.Accent.WithChromaAndTone(8, 75)),
		A1: option.Must(base.Main.WithChromaAndTone(16, 22)),
		A0: base.Accent,

		PrimaryButtonText: calcTextColor(base.Interactive),

		BannerErrorBackground: "#F6d2de",
		BannerErrorText:       "#FF543E",
		BannerInfoBackground:  "#1c3b12",
		BannerInfoText:        "#55ff3e",
	}

	withSemanticColors(&c)

	return c
}

func TrueLightMode(base BaseColors) ui.Colors {
	c := ui.Colors{
		M9: option.Must(base.Main.WithChromaAndTone(8, 92)),
		M8: option.Must(base.Main.WithChromaAndTone(8, 10)),
		M7: option.Must(base.Main.WithChromaAndTone(8, 60)),
		M6: option.Must(base.Main.WithChromaAndTone(8, 90)),
		M5: option.Must(base.Main.WithChromaAndTone(8, 70)),
		M4: option.Must(base.Main.WithChromaAndTone(8, 94)),
		M3: option.Must(base.Main.WithChromaAndTone(8, 90)),
		M2: option.Must(base.Main.WithChromaAndTone(6, 96)),
		M1: option.Must(base.Main.WithChromaAndTone(4, 98)),
		M0: base.Main,

		I1: option.Must(base.Interactive.WithChromaAndTone(16, 22)),
		I0: base.Interactive,

		A2: option.Must(base.Accent.WithChromaAndTone(8, 75)),
		A1: option.Must(base.Main.WithChromaAndTone(16, 22)),
		A0: base.Accent,

		PrimaryButtonText: calcTextColor(base.Interactive),

		BannerErrorBackground: "#F6d2de",
		BannerErrorText:       "#FF543E",
		BannerInfoBackground:  "#1c3b12",
		BannerInfoText:        "#55ff3e",
	}

	withSemanticColors(&c)

	return c
}

func withSemanticColors(c *ui.Colors) {
	c.Error = "#F47954"
	c.Warning = "#F7A823"
	c.Informative = "#17428C"
	c.Good = "#2BCA73"
	c.Disabled = "#E2E2E2"
	c.DisabledText = "#848484"
}

func calcTextColor(background ui.Color) ui.Color {
	c := xcolor.MustParseHex(string(background))
	if c.LumaBT709() < 128 {
		return "#FFFFFF"
	} else {
		return "#000000"
	}
}
