---
# Content is auto generated
# Manual changes will be overwritten!
title: Button
---
A basic clickable UI component used to trigger actions or events. There are three different kinds of Buttons:
PrimaryButton, SecondaryButton & TertiaryButton.

## Constructors
### Button
Button creates a new Button with the given style preset.

### PrimaryButton
PrimaryButton uses an internal preset to represent a primary button. See also FilledButton for a custom-colored
Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
```go
	PrimaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World")
```

![](/images/components/basic/buttons/primary-button.png)
```go
package main

import (
	"fmt"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

func main() {
	ui.PrimaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World").PreIcon(icons.SpeakerWave)

}

```

![](/images/components/basic/buttons/primary-button-with-pre-icon.png)

### SecondaryButton
SecondaryButton uses an internal preset to represent a secondary button. See also FilledButton for a custom-colored
Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
```go
	SecondaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World")
```

![](/images/components/basic/buttons/secondary-button.png)

### TertiaryButton
TertiaryButton uses an internal preset to represent a tertiary button. See also FilledButton for a custom-colored
Button. This may behave slightly different (but more correctly), due to optimizations of the frontend renderer.
```go
	TertiaryButton(func() {
		fmt.Println("Hello World")
	}).Title("Hello World")
```

![](/images/components/basic/buttons/tertiary-button.png)

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label used by screen readers for accessibility. |
| `Disabled(b bool)` |  |
| `Enabled(b bool)` | Enabled toggles whether the button is interactive. This has an effect only if a StylePreset is applied; otherwise it is ignored. |
| `Font(font Font)` | Font sets the font style for the button's text label. |
| `Frame(frame Frame)` | Frame sets the layout frame of the button, including size and positioning. |
| `FullWidth()` |  |
| `HRef(url core.URI)` | HRef sets the URL that the button navigates to when clicked if no action is specified. If both URL and Action are set, the URL takes precedence. This avoids another render cycle if the only goal is to navigate to a different page. It also avoids issues with browser which block async browser interactions like Safari. In fact, the [core.Navigation.Open] does not work properly on Safari. See also [TButton.Target]. |
| `ID(id string)` | ID assigns a unique identifier to the button, useful for testing or referencing. |
| `PostIcon(svg core.SVG)` | PostIcon sets the icon displayed after the text label. |
| `PreIcon(svg core.SVG)` | PreIcon sets the icon displayed before the text label. |
| `Preset(preset ButtonStyle)` | Preset applies a style preset to the button, controlling its appearance and behavior. |
| `Target(target string)` | Target sets the name of the browsing context, like _self, _blank, _ parent, _top. |
| `Title(text string)` | Title sets the text label displayed on the button. |
| `Visible(b bool)` | Visible controls the visibility of the button; setting false hides it. |
---

## Related
- [Frame](../../layout/frame/)

