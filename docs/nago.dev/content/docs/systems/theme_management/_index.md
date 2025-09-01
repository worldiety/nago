---
title: Theme Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/settings_management/galleries/theme_settings/overview.png"
galleryEdit:
  - src: "/images/systems/settings_management/galleries/theme_settings/edit.png"
  - src: "/images/systems/theme_management/galleries/edit.png"
---

The Theme Management system is responsible for **theme and Corporate Identity settings**.  
It allows configuration of logos, app icons, fonts, and color schemes, as well as legal and provider information.

{{< swiper name="galleryOverview" loop="false" >}}

## Functional areas
Theme Management offers the following key functions:

### Corporate Identity
- Define a **navigation bar logo** and an **app icon** (both for light and dark mode)
- Add a **slogan** or mission statement
- Configure **API contact information** (responsible provider, contact email, API documentation URL)

### Legal information
- Set external URLs or internal pages for:
    - Impressum
      Example `https://www.worldiety.de/impressum`
    - Privacy Policy
    - Terms & Conditions
    - User Agreement  
      Example: `http://localhost:3000/page/impressum`

{{< swiper name="galleryEdit" loop="false" >}}

### Visual customization
- Define **fonts** and **base colors** (main, interactive, accent) via code
- Separate colors for **light mode** and **dark mode**
- The system derives all additional color shades automatically from the base colors

### Example: Define base colors
```go
import (
    "log/slog"
    
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
    "go.wdy.de/nago/pkg/std"
)

themeManagement := std.Must(cfg.ThemeManagement())

// Read current colors
colors, err := themeManagement.UseCases.ReadColors(user.SU())
if err != nil {
    slog.Error("failed to read theme colors", slog.Any("error", err))
}

// Update base colors
err = themeManagement.UseCases.UpdateColors(user.SU(), theme.Colors{
    Dark: theme.BaseColors{
        Main:        ui.Color("#222222"),
        Interactive: ui.Color("#0055ff"),
        Accent:      ui.Color("#ff6600"),
    },
})
if err != nil {
    slog.Error("failed to update theme colors", slog.Any("error", err))
}
```

### Example: Define custom fonts
```go
import (
    _ "embed"
	
    "go.wdy.de/nago/application"
    "go.wdy.de/nago/application/settings"
    "go.wdy.de/nago/application/theme"
    "go.wdy.de/nago/pkg/std"
    "go.wdy.de/nago/presentation/ui"
)

//go:embed font/GloriaHallelujah-Regular.ttf
var fntGloria application.StaticBytes

//go:embed font/Silkscreen-Bold.ttf
var fntSilkscreenBold application.StaticBytes

uriGloria := cfg.Resource(fntGloria)
uriSilkBold := cfg.Resource(fntSilkscreenRegular)

cfgTheme := settings.ReadGlobal[theme.Settings](std.Must(cfg.SettingsManagement()).UseCases.LoadGlobal)
cfgTheme.Fonts.DefaultFont = "Gloria"
cfgTheme.Fonts.Faces = nil // clear whatever has been defined in the past
cfgTheme.Fonts.Faces = append(cfgTheme.Fonts.Faces,
    core.FontFace{
        Family: "Gloria",
        Source: uriGloria,
    },
    core.FontFace{
        Family: "Silk",
        Source: uriSilkBold,
        Weight: "bold",
    },
)
settings.WriteGlobal(std.Must(cfg.SettingsManagement()).UseCases.StoreGlobal, cfgTheme)

ui.Text("Default text in Gloria")
ui.Text("Custom text in Silk").Font(ui.Font{Name: "Silk"})
```

## Dependencies
**Requires:**
- [Settings Management](../settings_management/) for storing global theme configurations

If this is not already active, it will be enabled automatically when Theme Management is activated.

**Is required by:**
- none

{{< callout type="info" >}}
Theme Management is applied automatically on application startup via `Run()`
{{< /callout >}}

## Activation
This system is activated via:

```go
std.Must(cfg.ThemeManagement())
```
```go
themeManagement := std.Must(cfg.ThemeManagement())
```

