---
# Content is auto generated
# Manual changes will be overwritten!
title: Picker
---
The picker component is classic dropdown menu.

## Constructors
### Picker
Picker takes the given slice and state to represent the selection. Internally, it uses deep equals, to determine
the unique set of selected elements and coordinate that with the UI state.
```go
package main

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/picker"
)

func main() {
	type Person struct {
		Name string
		Age  int
	}

	persons := []Person{
		{
			Name: "John",
			Age:  20,
		},
		{
			Name: "Jane",
			Age:  30,
		},
	}

	selected := core.AutoState[[]Person](nil)
	picker.Picker[Person]("Ich bin ein picker", persons, selected)
}

```

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` |  |
| `Border(border ui.Border)` |  |
| `DetailView(detailView core.View)` | DetailView is optional and placed between the picker section and the button footer. |
| `Dialog()` | Dialog returns the dialog view as if pressed on the actual button. |
| `DialogPresented()` |  |
| `Disabled(disabled bool)` |  |
| `ErrorText(text string)` |  |
| `Frame(frame ui.Frame)` |  |
| `ItemPickedRenderer(fn func([]T) core.View)` | ItemPickedRenderer can be customized to return a non-text view for the given T. This is shown within the selected window for the currently selected items. |
| `ItemRenderer(fn func(T) core.View)` | ItemRenderer can be customized to return a non-text view for the given T. This is shown within the picker popup. If fn is nil, the default fallback rendering will be applied. |
| `MultiSelect(mv bool)` | MultiSelect is by default false. |
| `Padding(padding ui.Padding)` |  |
| `QuickFilterSupported(flag bool)` | QuickFilterSupported sets the quick-filter-support and if true and values contains more than 10 items, the quick filter is shown. Default is true. |
| `SelectAllSupported(flag bool)` | SelectAllSupported sets the select-all-support and if true and multiSelect is enabled, a checkbox to select all is shown. Default is true. |
| `SupportingText(text string)` |  |
| `Title(title string)` |  |
| `Visible(visible bool)` |  |
| `WithDialogPresented(state *core.State[bool])` |  |
| `WithFrame(fn func(ui.Frame) ui.Frame)` |  |
| `pickerTable()` |  |
| `syncCheckboxStates(state *core.State[[]T])` |  |
| `syncCurrentSelectedState()` |  |
---
## Related

- [Frame](../../layout/frame/)
