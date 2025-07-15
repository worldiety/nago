---
# Content is auto generated
# Manual changes will be overwritten!
title: Text
---
This component displays a piece of text with rich styling and interaction options. It can be used for static labels, links, or interactive elements with click actions.

## Constructors
### Link
Link performs a best guess based on the given href. If the href starts with http or https
the window will perform an Open call. Otherwise, a local forward navigation is applied.
```go
	Link(nil, "Nago Docs", "https://www.nago-docs.com", "_blank")
```

![](/images/components/basic/text/link-example.png)

### LinkWithAction
```go
	LinkWithAction("Nago Docs", func() {
		fmt.Printf("Nago is easy to use")
	})
```

![](/images/components/basic/text/link-example.png)

### MailTo
```go
	MailTo(nil, "Worldiety", "info@worldiety.de")
```

![](/images/components/basic/text/mail-to-example.png)

### Text
```go
package main

import (
	"fmt"
	"go.wdy.de/nago/presentation/ui"
)

func main() {
	ui.Text("hello world").
		Action(func() {
			fmt.Print("Nago is easy to use")
		}).
		Underline(true).
		Color("#eb4034").
		Border(ui.Border{}.Width("2px").Color("#4287f5"))
}

```

![](/images/components/basic/text/text-with-methods-example.png)

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the label of the text. The content of the label is also displayed in the tooltip that appears when you hover over the Text. |
| `Action(f func())` | Action executes the function when the component is clicked. |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the color of the background. |
| `Border(border Border)` | Border draws a Border around the component. It's used to set the Border width, color and radius. Fore more information also have a look at the Border component. |
| `Color(color Color)` | Color sets the Color of the font. |
| `FocusedBorder(border Border)` | FocusedBorder sets the Border width, color and radius when the component is focused. |
| `Font(font Font)` | Font sets the size, style and width of the Text. For more information also have a look at Font. |
| `Frame(frame Frame)` | Frame sets the width, minWidth, maxWidth, height, minHeight and maxHeight. |
| `FullWidth()` | FullWidth sets the width to 100%. |
| `HoveredBorder(border Border)` | HoveredBorder sets the Border width, color and radius when component is hovered. |
| `LineBreak(lb bool)` | LineBreak de-/activates line breaking in between the Text. |
| `Padding(padding Padding)` | Padding sets a top, right, bottom and left spacing. |
| `PressedBorder(border Border)` | PressedBorder sets the Border width, color and radius when the component is clicked. |
| `TextAlignment(align TextAlignment)` | TextAlignment sets the position of the Text. For more information also have a look at TextAlignment. |
| `Underline(b bool)` | Underline underlines the Text. |
| `Visible(visible bool)` | Visible decides whether a text is shown. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame sets width, minWidth, maxWidth, height, minHeight and maxHeight using a function. |
---

## Related
- [Frame](../../layout/frame/)

## Tutorials
- [tutorial-01-helloworld](../../../examples/tutorial-01-helloworld)
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
- [tutorial-54-codeeditor](../../../examples/tutorial-54-codeeditor)
