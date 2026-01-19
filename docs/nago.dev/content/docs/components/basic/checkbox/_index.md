---
# Content is auto generated
# Manual changes will be overwritten!
title: Checkbox
---
It allows users to toggle between checked and unchecked states,
optionally binding to external state. The checkbox can be disabled,
hidden, or assigned a unique identifier for reference.

## Constructors
### Checkbox
Checkbox represents a user interface element which spans a visible area to click or tap from the user.
Use it for controls, which do not cause an immediate effect. See also [Toggle].

---
## Methods
| Method | Description |
|--------| ------------|
| `Disabled(disabled bool)` | Disabled enables or disables user interaction with the checkbox. |
| `ID(id string)` | ID assigns a unique identifier to the checkbox, useful for testing or referencing. |
| `InputChecked(input *core.State[bool])` | Deprecated: use InputValue InputChecked binds the checkbox to an external boolean state, allowing it to be controlled from outside the component. |
| `InputValue(input *core.State[bool])` | InputValue binds the checkbox to an external boolean state, allowing it to be controlled from outside the component. |
| `Visible(v bool)` | Visible controls the visibility of the checkbox; setting false hides it. |
---

