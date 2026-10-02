---
title: Colors
---

`Colors` is the color set of the Nago design system. It holds the actual values behind the theme
[color names](../color/) like `M4` or `I0` for one color scheme (light or dark). Usually the theme system computes
it from the main, accent and interactive colors of your theme; read the current values with
`core.Colors[Colors](wnd)`.

```go
oraColors := core.Colors[Colors](wnd)

return Text("hello world").
	Color(oraColors.I0).
	BackgroundColor(oraColors.M0)
```

`Colors` implements `core.ColorSet`. You can define your own color sets the same way: a struct with public `Color`
fields, a unique namespace and default values. Register them with `cfg.ColorSet(scheme, set)` on the configurator,
see the [Colors tutorial](/docs/examples/tutorial-08-colors/).

## Fields

| Field | Description |
|-------|-------------|
| `M0` to `M9` | Main source color and its shades. |
| `A0` to `A2` | Accent source color and its variants. |
| `I0`, `I1` | Interactive source color and the button color. |
| `Error`, `Warning`, `Good`, `Informative` | Semantic colors. |
| `Disabled`, `DisabledText` | Disabled input areas and text on them. |
| `PrimaryButtonText` | Text color on primary buttons. |
| `Banner…Background`, `Banner…Text` | Background and text colors of error, info, warning and success banners. |

## Methods

| Method | Description |
|--------|-------------|
| `Default(scheme core.ColorScheme) core.ColorSet` | Returns fallback colors; the real defaults are computed by the theme system. |
| `Namespace() core.NamespaceName` | Returns the reserved namespace `ora`. |
| `Valid() bool` | Reports whether all main, accent, interactive, semantic and disabled colors are set. |

## Related

- [Color](../color/)
- Tutorials: [Colors](/docs/examples/tutorial-08-colors/), [Theme](/docs/examples/tutorial-55-theme/)
