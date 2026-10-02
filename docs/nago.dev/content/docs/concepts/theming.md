---
title: Theming
weight: 7
---

Nago applications support a light and a dark color scheme out of the box. The frontend follows the system
preference of the user, and you can offer a switch with `ui.ThemeSwitcher` or request a scheme with
`wnd.SetColorScheme(core.Dark)`, see [tutorial-102-theme-switch](/docs/examples/tutorial-102-theme-switch/).

## Color variables

Instead of fixed colors, components use color variables which the frontend resolves for the active scheme. They
are defined as constants of type `ui.Color` in `go.wdy.de/nago/presentation/ui`:

| Variables       | Meaning                                                                                     |
|-----------------|---------------------------------------------------------------------------------------------|
| `M0` … `M9`     | derived from the main color: backgrounds, containers, cards, lines and text                 |
| `I0`, `I1`      | derived from the interactive color: buttons and other interactive elements                  |
| `A0`, `A1`      | derived from the accent color: highlights, progress bars, borders                           |
| `SE0`, `SW0`, `SG0`, `SV0` | semantic colors for error, warning, good and informative                         |

Use them like any other color:

```go
ui.Text("hello").Color(ui.M8).BackgroundColor(ui.M2)
```

A `ui.Color` can also be a hex value like `"#ff0000"`, which is the same in both schemes. Prefer the variables,
so that your application looks right in both schemes and follows the theme.

To read the actual values of the current scheme in Go, use `core.Colors[ui.Colors](wnd)`.

## Theme colors

All variables are calculated from three base colors: main, interactive and accent. The theme system stores them
for both schemes in the global settings and applies changes immediately:

```go
themes := option.Must(cfg.ThemeManagement())

if !option.Must(themes.UseCases.HasColors(user.SU())) {
	base := theme.BaseColors{Main: "#1b8c30", Interactive: "#17428c", Accent: "#f7a823"}

	option.MustZero(themes.UseCases.UpdateColors(user.SU(), theme.Colors{
		Dark:  themes.UseCases.Calculations.DarkMode(base),
		Light: themes.UseCases.Calculations.LightMode(base),
	}))
}
```

The check with `HasColors` keeps colors which were changed at runtime. See
[tutorial-55-theme](/docs/examples/tutorial-55-theme/) and [Theme management](/docs/systems/theme_management/),
which also manages logos, app icons and legal information.

## Custom color sets

For colors of your own, define a struct with flat, exported fields of type `ui.Color` which implements
`core.ColorSet` (`Default(scheme)` and `Namespace()`), register values with `cfg.ColorSet(scheme, set)` and read
them with `core.Colors[MyColors](wnd)`. See [tutorial-08-colors](/docs/examples/tutorial-08-colors/).

## Fonts

Fonts are part of the theme settings, too. Embed font files, register them with `cfg.Resource` and add them as
`core.FontFace` to the theme settings, see [tutorial-60-customfont](/docs/examples/tutorial-60-customfont/) and
[tutorial-111-typography](/docs/examples/tutorial-111-typography/).
