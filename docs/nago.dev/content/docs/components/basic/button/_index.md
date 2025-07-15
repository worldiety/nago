---
# Content is auto generated
# Manual changes will be overwritten!
title: Button
---
A basic clickable UI component used to trigger actions or events. There are three different kinds of Buttons:
PrimaryButton, SecondaryButton & TertiaryButton.

## Constructors
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
| `AccessibilityLabel(label string)` |  |
| `Enabled(b bool)` | Enabled has only an effect for StylePreset otherwise it is ignored. |
| `Font(font Font)` |  |
| `Frame(frame Frame)` |  |
| `ID(id string)` |  |
| `PostIcon(svg core.SVG)` |  |
| `PreIcon(svg core.SVG)` |  |
| `Preset(preset StylePreset)` |  |
| `Title(text string)` |  |
| `Visible(b bool)` |  |
---
## Related
- [Frame](../../layout/frame/)
- [Button](../../basic/button/)

