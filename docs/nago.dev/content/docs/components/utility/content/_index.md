---
title: Content
---

`editor.Content` of package [`editor`](https://github.com/worldiety/nago/tree/main/presentation/ui/editor) is the
main area of an editor [Screen](../../composite/screen/). It wraps a single view in a scrollable container, either
filling the whole area or styled as a centered page.

```go
editor.Screen("Document").
	Content(editor.Content(Text("The document")).Style(editor.ContentPage))
```

## Constructors

| Constructor | Description |
|-------------|-------------|
| `editor.Content(view core.View) TContent` | Creates a new TContent with the given view and sets it visible by default. |

## Methods

| Method | Description |
|--------|-------------|
| `Style(style ContentStyle) TContent` | Applies `ContentFull` (the default, a scrollable full-size area) or `ContentPage` (a scrollable, styled page). |

## Related

- [Screen](../../composite/screen/), [Tool Window](../../composite/tool_window/), [ScrollView](../../layout/scroll_view/)
