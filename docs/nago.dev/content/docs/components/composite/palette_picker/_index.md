---
# Content is auto generated
# Manual changes will be overwritten!
title: Palette Picker
---
This component allows users to select a color
from a predefined palette. It is typically used in design tools or
configuration interfaces where color choices are limited to a fixed set.

## Constructors
### PalettePicker
```go
package main

import abc "go.wdy.de/nago/presentation/ui/colorpicker"

func main() {
	abc.PalettePicker("Colorpicker", abc.DefaultPalette)
}

```
```go
package main

import (
	"fmt"
	abc "go.wdy.de/nago/presentation/ui/colorpicker"
)

func main() {
	fmt.Println("klappt")
	abc.PalettePicker("2. Example", abc.DefaultPalette)
}

```

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` |  |
| `Border(border ui.Border)` |  |
| `Dialog(pickerPresented *core.State[bool])` | Dialog returns the dialog view as if pressed on the actual button. |
| `Disabled(disabled bool)` |  |
| `ErrorText(text string)` |  |
| `Frame(frame ui.Frame)` |  |
| `Padding(padding ui.Padding)` |  |
| `State(state *core.State[ui.Color])` | State attaches the given state to the interaction process of selecting a value. A nil state signals read-only. |
| `SupportingText(text string)` |  |
| `Title(title string)` |  |
| `Value(color ui.Color)` | Value sets the selected value. An empty Color selects none. |
| `Visible(visible bool)` |  |
| `WithFrame(fn func(ui.Frame) ui.Frame)` |  |
| `pickerTable()` |  |
---
## Related
- [Palette Picker](../../composite/palette_picker/)

