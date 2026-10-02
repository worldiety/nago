---
title: Page
---

A page of package [`tabs`](https://github.com/worldiety/nago/tree/main/presentation/ui/tabs) is one entry of
[Tabs](../tabs/): a title, an optional icon and a function which creates the content. A disabled page is shown
but cannot be selected.

![Tabs with pages](tabs.webp)

```go
return tabs.Tabs(
	tabs.Page("Overview", func() core.View {
		return Text("The content of the overview page.")
	}).Icon(icons.Home),
	tabs.Page("Settings", func() core.View {
		return Text("The content of the settings page.")
	}).Icon(icons.Cog6Tooth),
	tabs.Page("Archive", func() core.View {
		return Text("Not available yet.")
	}).Disabled(true),
).InputValue(core.AutoState[int](wnd)).
	Frame(Frame{Width: L560})
```

## Constructors

| Constructor | Description |
|-------------|-------------|
| `tabs.Page(title string, body func() core.View) TPage` | Creates a new tab page with the given title and content body. |

## Methods

| Method | Description |
|--------|-------------|
| `Disabled(disabled bool) TPage` | Marks the page as disabled, making it visible but not selectable. |
| `Icon(ico core.SVG) TPage` | Sets the icon displayed next to the page title. |

## Related

- [Tabs](../tabs/)
