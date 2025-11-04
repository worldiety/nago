---
# Content is auto generated
# Manual changes will be overwritten!
title: Toggle
---
This component represents a switch-like control (on/off) without a label. It is intended for immediate activation or deactivation of features.

## Constructors
### Toggle
Toggle is just a kind of checkbox without a label. However, a toggle shall be used for immediate activation
functions. In contrast to that, use a checkbox for form things without an immediate effect.

---
## Methods
| Method | Description |
|--------| ------------|
| `Disabled(disabled bool)` | Disabled enables or disables interaction with the toggle. |
| `InputChecked(input *core.State[bool])` | InputChecked binds the toggle to an external boolean state for two-way data binding. |
| `Visible(v bool)` | Visible controls the visibility of the toggle; false hides it from the UI. |
---

