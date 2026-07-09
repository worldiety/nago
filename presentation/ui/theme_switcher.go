// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"slices"

	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/material/outlined"
	"go.wdy.de/nago/presentation/proto"
)

var (
	StrTheme       = core.DefaultStr("theme-switcher.theme", "Color scheme", "Farbschema")
	StrThemeLight  = core.DefaultStr("theme-switcher.theme.light", "Light", "Hell")
	StrThemeDark   = core.DefaultStr("theme-switcher.theme.dark", "Dark", "Dunkel")
	StrThemeSystem = core.DefaultStr("theme-switcher.theme.system", "System", "System")
)

// TThemeSwitcher is a component to display a theme switching ui.
// It displays a dropdown menu anchored to a specific view.
type TThemeSwitcher struct {
	anchor core.View // view the menu is anchored to
	frame  Frame     // layout frame for sizing and positioning
}

// ThemeSwitcher creates a new theme switching menu with the given anchor.
func ThemeSwitcher(anchor core.View) TThemeSwitcher {
	return TThemeSwitcher{
		anchor: anchor,
	}
}

// Frame sets the layout frame of the menu.
func (c TThemeSwitcher) Frame(frame Frame) TThemeSwitcher {
	c.frame = frame
	return c
}

// Render builds and returns the protocol representation of the menu.
func (c TThemeSwitcher) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	themes := []string{core.Light.String(), core.Dark.String(), core.System.String()}
	themeByIdx := func(idx int) core.ColorScheme {
		switch themes[idx] {
		case core.Light.String():
			return core.Light
		case core.Dark.String():
			return core.Dark
		default:
			return core.System
		}
	}

	themeIndex := slices.Index(themes, wnd.Info().ColorScheme.String())
	stateTheme := AutoRadioStateGroup(wnd, "stateTheme", len(themes)).InitIndex(themeIndex)

	if themeIndex != stateTheme.SelectedIndex() {
		wnd.SetColorScheme(themeByIdx(stateTheme.SelectedIndex()))
	}

	stateTheme.Observe(func(idx int) {
		switch themes[idx] {
		case core.Light.String():
			wnd.SetColorScheme(core.Light)
		case core.Dark.String():
			wnd.SetColorScheme(core.Dark)
		default:
			wnd.SetColorScheme(core.System)
		}
	})

	return &proto.Menu{
		Anchor: render(ctx, c.anchor),
		Groups: []proto.MenuGroup{
			{
				CustomContent: VStack(
					VStack(
						ImageIcon(icons.Palette).FillColor(M8),
						Text(StrTheme.Get(wnd)),
					).FullWidth(),
					VStack(
						Each2(stateTheme.All(), func(idx int, checked *core.State[bool]) core.View {
							return RadioButtonField(getLabelFromTheme(wnd, themes[idx]), &stateTheme, idx)
						})...,
					).Alignment(Leading).FullWidth().Padding(Padding{}.Vertical(L8)),
				).FullWidth().Padding(Padding{}.All(L8)).Render(ctx),
			},
		},
		Frame:  c.frame.ora(),
		Offset: proto.Length(L8),
	}
}

func getLabelFromTheme(wnd core.Window, theme string) string {
	switch theme {
	case core.Light.String():
		return StrThemeLight.Get(wnd)
	case core.Dark.String():
		return StrThemeDark.Get(wnd)
	case core.System.String():
		return StrThemeSystem.Get(wnd)
	default:
		return "?"
	}
}
