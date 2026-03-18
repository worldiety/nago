---
# Content is auto generated
# Manual changes will be overwritten!
title: Stack
---
It is responsive and can switch between [HStack] and [VStack] during rendering.

## Constructors
### HSwitcher
HSwitcher is a fixed horizontal variant of Switcher

### Switcher
Switcher is a responsive variant which decides between HSwitcher and VSwitcher.

### VSwitcher
VSwitcher is a fixed vertical variant of Switcher

---
## Methods
| Method | Description |
|--------| ------------|
| `Append(pages ...)` | Append adds more pages to the switcher |
| `ContentNoPadding()` | ContentNoPadding sets whether the content part should use the prestyled padding |
| `DynamicHeight()` | DynamicHeight sets the switcher to dynamically change its height by the active page |
| `Frame(frame ui.Frame)` | Frame sets the switcher's frame |
| `FullWidth()` | FullWidth sets the switcher's frame to full width |
| `ID(id string)` | ID assigns a unique identifier to the switcher |
| `InputValue(input *core.State[string])` | InputValue binds the switcher to an external string state, allowing it to be controlled from outside the component. |
| `Layout(layout SwitcherLayout)` | Layout sets the switcher's layout |
| `With(fn func(switcher TSwitcher) TSwitcher)` | With applies a transformation function to the switcher itself and returns the result. Useful for chaining configuration in a functional style. |
---

