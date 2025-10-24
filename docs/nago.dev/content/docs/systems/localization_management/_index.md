---
title: Localization Management
galleryOverview:
  - src: "/images/systems/shared/admin_center.png"
  - src: "/images/systems/localization_management/galleries/overview/admin_center.png"
galleryEdit:
  - src: "/images/systems/localization_management/galleries/edit/admin_center.png"
  - src: "/images/systems/localization_management/galleries/edit/nago_domain.png"
  - src: "/images/systems/localization_management/galleries/edit/common_overview.png"
  - src: "/images/systems/localization_management/galleries/edit/action_overview.png"
  - src: "/images/systems/localization_management/galleries/edit/add_translation.png"
galleryLanguages:
  - src: "/images/systems/localization_management/galleries/languages/admin_center.png"
  - src: "/images/systems/localization_management/galleries/languages/add.png"
---

The Localization Management system provides **internationalization (i18n)** capabilities for Nago-based applications.  
It enables multilingual support by allowing both developers and users to define and manage translations for UI texts, messages, and labels.  
The system is built on top of [github.com/worldiety/i18n](https://github.com/worldiety/i18n), which offers a fast, developer-centric localization API optimized for runtime translation management.

## Functional areas
Localization Management provides the following core functions:

### Translation directory
- View and navigate the hierarchy of translation keys based on their logical namespaces (e.g. `mydomain.example`)
- Overview of total and missing translations per section
- Direct access to localized message editing

{{< swiper name="galleryOverview" loop="false" >}}

### Editing translations
- View and edit translations for all supported languages directly in the Admin Center
- Update existing messages or add missing translations without redeploying the application
- Changes are stored persistently and immediately reflected in the running app

{{< swiper name="galleryEdit" loop="false" >}}

### Language management
- Add new languages dynamically through the Admin Center UI
- Default languages: **English** and **German** (preloaded for all standard systems and components)
- Additional languages are added at runtime and initialized automatically in the i18n resource bundles

{{< swiper name="galleryLanguages" loop="false" >}}

### Developer integration
Developers define all translatable strings in Go code using the `i18n API`. 

#### Example: Must String
```go
var StrHelloWorld = i18n.MustString(
	"mydomain.example.hello_world",
	i18n.Values{
		language.English: "Hello World",
		language.German:  "Hallo Welt",
	},
)
```

```go
cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
    return ui.VStack(
		ui.Text(StrHelloWorld.Get(wnd)), 
	).FullWidth()
})
```

#### Example: Must Quantity String
```go
var LabelXItems = i18n.MustQuantityString(
    "nago.common.label.x_items",
	i18n.QValues{
        language.English: i18n.Quantities{
            One:   "{x} item",
			Other: "{x} items",
		},
        language.German: i18n.Quantities{
			One:   "{x} Element", 
			Other: "{x} Elemente",
		},
	},
	i18n.LocalizationHint("Shown when listing items, e.g. '3 items'"), 
)
```
  
```go
cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
    return ui.VStack(
	    ui.Text(LabelXItems.Get(wnd, 7,  i18n.String("x", "seven"))),
    ).FullWidth()
})
```

- The Admin Center automatically exposes these keys for translation and correction
- Translations are cached and loaded **efficiently** for **O(1) lookup performance**

{{< callout type="info" >}}
Localization Management is designed to empower end users to fix or extend translations directly —
reducing dependency on professional translation workflows and speeding up internationalization cycles.
{{< /callout >}}

## Dependencies
**Requires:**
- None

**Is required by:**
- None

## Activation
This system is activated via:

```go
std.Must(cfglocalization.Enable(cfg))
```
```go
localizationManagement := std.Must(cfglocalization.Enable(cfg))
```

