---
title: Localization
weight: 8
---

Nago translates texts with the package [github.com/worldiety/i18n](https://github.com/worldiety/i18n). The
built-in systems ship English and German texts.

## Language of a window

The frontend sends the preferred languages of the browser (`Accept-Language`) when it connects. Nago picks the
best matching translation bundle for the window. `wnd.Locale()` returns the negotiated language tag and
`wnd.Bundle()` the bundle; a subject provides its language and bundle as well, so use cases can localize, too.

## Declaring strings

Declare your strings once at package level. Each declaration registers the key and its translations in
`i18n.Default` and returns a handle:

```go
var (
	StrHelloWorld = i18n.MustString(
		"com.example.myapp.hello_world",
		i18n.Values{language.English: "hello world", language.German: "Hallo Welt"},
	)

	StrHelloX = i18n.MustVarString(
		"com.example.myapp.hello_x",
		i18n.Values{language.English: "hello {name}", language.German: "Hallo {name}"},
		i18n.LocalizationHint("greets the user on the start page"),
		i18n.LocalizationVarHint("name", "the first name of the user"),
	)
)
```

`language` is `golang.org/x/text/language`. `i18n.MustEnglish` declares an English-only string as a starting
point. Keys must be unique; declaring a key twice panics.

Resolve a string with `Get`, passing the window or a subject and, for variables, their values:

```go
ui.Text(StrHelloWorld.Get(wnd))
ui.Text(StrHelloX.Get(wnd, i18n.String("name", wnd.Subject().Firstname())))
```

The package `go.wdy.de/nago/application/localization/rstring` contains predefined standard texts, e.g.
`rstring.ActionSave`, which you can reuse.

## Translating at runtime

`cfglocalization.Enable(cfg)` from `go.wdy.de/nago/application/localization/cfg` adds pages to the admin center
where administrators can review all keys, find missing translations, add languages and edit texts without a new
release. See [tutorial-74-localization](/docs/examples/tutorial-74-localization/) and
[Localization management](/docs/systems/localization_management/).
