// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

// Color is a color like "#ff0000" or a color of the theme like M0 or ColorAccent, which adapts to the light and the
// dark mode.
type Color string

const (
	// M0 is the main color of the background.
	M0 Color = "M0"
	// M1 is a variable name usually used for the background.
	M1 Color = "M1"
	// M2 is a variable name usually used as the background for first level container.
	M2 Color = "M2"
	// M3 is a variable name usually used for a card bottom area.
	M3 Color = "M3"
	// M4 is a variable name usually used for a card body area.
	M4 Color = "M4"
	// M5 is a variable name usually used for Line / Dot on SC.
	M5 Color = "M5"
	// M6 is a variable name usually used for hovered containers.
	M6 Color = "M6"
	// M7 is a variable name usually used for Text or muted icons.
	M7 Color = "M7"
	// M8 is the main color of the text.
	M8 Color = "M8"
	// M9 is a variable name usually used for card Top area.
	M9 Color = "M9"
	// A0 is the accent color.
	A0 Color = "A0"
	// A1 is a variable name usually used for card progress bars, H2 or Borders.
	A1 Color = "A1"
	// I0 is the source interactive color name.
	I0 Color = "I0"
	// I1 is a variable name usually used for buttons.
	I1 Color = "I1"
	// SE0 is the source error color variable name.
	SE0 Color = "SE0"
	// SW0 is the source warning color variable name.
	SW0 Color = "SW0"
	// SG0 is the source good color variable name.
	SG0 Color = "SG0"
	// SV0 is the source informative color variable name.
	SV0 Color = "SV0"
	// SI0 is the source disabled input color variable name.
	SI0 Color = "SI0"
	// ST0 is the source disabled text color variable name.
	ST0 Color = "ST0"
)

// Aliases of the colors above, which name their purpose.
const (
	// ColorCardBody is the conventional card body color derived from the main color, see M4.
	ColorCardBody Color = "M4"
	// ColorCardTop is the color of the top area of a card, see M9.
	ColorCardTop Color = "M9"
	// ColorCardFooter is the color of the bottom area of a card, see M3.
	ColorCardFooter Color = "M3"
	// ColorContainer is the color of hovered containers, see M6.
	ColorContainer Color = "M6"
	// ColorBackground is the color of the background, see M1.
	ColorBackground Color = "M1"
	// ColorAccent is the exact accent color, see A0.
	ColorAccent Color = "A0"
	// ColorInputBorder is the color of the border of input elements like a text field, see M8.
	ColorInputBorder Color = "M8"
	// ColorText is the color of the text, see M8.
	ColorText Color = "M8"
	// ColorLine is the conventional color for a line derived from the main color, see M5.
	ColorLine Color = "M5"
	// ColorError is the error color, see SE0.
	ColorError Color = "SE0"
	// ColorBlack is black.
	ColorBlack Color = "#000000"
	// ColorWhite is white.
	ColorWhite Color = "#ffffff"
	// ColorIcons is the color of icons and of the default text, see M8.
	ColorIcons Color = "M8"
	// ColorIconsMuted is the color of muted icons, see M7.
	ColorIconsMuted Color = "M7"
	// ColorInteractive is the interactive color, see I0.
	ColorInteractive Color = "I0"
	// ColorSemanticGood is the color of a good state, see SG0.
	ColorSemanticGood Color = "SG0"
	// ColorSemanticWarn is the color of a warning, see SW0.
	ColorSemanticWarn Color = "SW0"
	// ColorSemanticError is the color of an error, see SE0.
	ColorSemanticError Color = "SE0"
	// ColorBannerErrorBackground is the background of an error banner.
	ColorBannerErrorBackground Color = "CBEB"
	// ColorBannerErrorText is the text of an error banner.
	ColorBannerErrorText Color = "CBET"
	// ColorBannerInfoBackground is the background of an info banner.
	ColorBannerInfoBackground Color = "CBIB"
	// ColorBannerInfoText is the text of an info banner.
	ColorBannerInfoText Color = "CBIT"
	// ColorBannerWarningBackground is the background of a warning banner.
	ColorBannerWarningBackground Color = "CBWB"
	// ColorBannerWarningText is the text of a warning banner.
	ColorBannerWarningText Color = "CBWT"
	// ColorBannerSuccessBackground is the background of a success banner.
	ColorBannerSuccessBackground Color = "CBSB"
	// ColorBannerSuccessText is the text of a success banner.
	ColorBannerSuccessText Color = "CBST"
)
