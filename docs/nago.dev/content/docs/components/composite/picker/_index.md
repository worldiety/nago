---
# Content is auto generated
# Manual changes will be overwritten!
title: Picker
---
It displays a list of values and lets users choose one or multiple items,
with optional "Select all" and quick-filtering support. Rendering of both the
selected summary and the selectable rows is customizable via callbacks. The picker can bind to external selection state or manage its own, and it
can be presented in a dialog with configurable options.

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
| `DialogOptions(opts ...)` |  |
| `DialogPresented()` |  |
| `Disabled(disabled bool)` |  |
| `ErrorText(text string)` |  |
| `Frame(frame ui.Frame)` |  |
| `ItemPickedRenderer(fn func([]T) core.View)` | ItemPickedRenderer can be customized to return a non-text view for the given T. This is shown within the selected window for the currently selected items. |
| `ItemRenderer(fn func(T) core.View)` | Deprecated: ItemRenderer can be customized to return a non-text view for the given T. This is shown within the picker popup. If fn is nil, the default fallback rendering will be applied. |
| `ItemRenderer2(fn func(wnd core.Windowitem Tstate *core.State[bool]) core.View)` | ItemRenderer2 can be customized to return a non-text view for the given T. This is shown within the picker popup. If fn is nil, the default fallback rendering will be applied. |
| `MultiSelect(mv bool)` | MultiSelect is by default false. |
| `Padding(padding ui.Padding)` |  |
| `QuickFilterSupported(flag bool)` | QuickFilterSupported sets the quick-filter-support and if true and values contains more than 10 items, the quick filter is shown. Default is true. |
| `SelectAllSupported(flag bool)` | SelectAllSupported sets the select-all-support and if true and multiSelect is enabled, a checkbox to select all is shown. Default is true. |
| `SupportingText(text string)` |  |
| `Title(title string)` |  |
| `Visible(visible bool)` |  |
| `WithDialogPresented(state *core.State[bool])` |  |
| `WithFrame(fn func(ui.Frame) ui.Frame)` |  |
| `pickerTable(wnd core.Window)` |  |
| `syncCheckboxStates(state *core.State[[]T])` |  |
| `syncCurrentSelectedState()` |  |
---

