---
title: Checkbox
---

A checkbox lets the user switch a boolean value on or off without an immediate effect, e.g. in a form which is
submitted later. Bind it to a `*core.State[bool]` with `InputValue` and pass the current value to the constructor.
In forms you usually want the [Checkbox Field](../checkbox_field/), which adds a label. For settings which take
effect immediately, use a [Toggle](../toggle/).

![Checkbox](checkbox.webp)

```go
checked := core.AutoState[bool](wnd).Init(func() bool { return true })

HStack(
	Checkbox(checked.Get()).InputValue(checked),
	Checkbox(false),
	Checkbox(true).Disabled(true),
).Gap(L16)
```

## Constructors

| Constructor | Description |
|---|---|
| `func Checkbox(checked bool) TCheckbox` | Creates a checkbox with the given value. |

## Methods

| Method | Description |
|---|---|
| `Disabled(disabled bool) TCheckbox` | Disabled enables or disables user interaction with the checkbox. |
| `ID(id string) TCheckbox` | ID assigns a unique identifier to the checkbox, useful for testing or referencing. |
| `InputChecked(input *core.State[bool]) TCheckbox` | Deprecated: use InputValue. |
| `InputValue(input *core.State[bool]) TCheckbox` | InputValue binds the checkbox to an external boolean state, allowing it to be controlled from outside the component. |
| `Name(name string) TCheckbox` | Name defines the name of the checkbox. |
| `Visible(v bool) TCheckbox` | Visible controls the visibility of the checkbox; setting false hides it. |

## Related

- [Checkbox Field](../checkbox_field/)
- [Toggle](../toggle/)
- [Radio Button](../radio_button/)
- Tutorials: [Checkbox](/docs/examples/tutorial-14-checkbox/)
