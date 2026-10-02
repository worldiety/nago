---
title: Keyboard Options
---

Keyboard options are hints for the virtual keyboard of mobile devices: which keyboard type to show and whether to
capitalize and auto-correct. Pass them to a [text field](../../basic/text_field/) with its `KeyboardOptions`
method.

```go
TextField("E-mail", email.Get()).
	InputValue(email).
	KeyboardOptions(KeyboardOptions().
		KeyboardType(KeyboardEMail).
		Capitalization(false).
		AutoCorrectEnabled(false))
```

Keyboard types are hints only. Users may use other keyboards, paste text or send anything to your server, so
always validate the input.

## Constructors

| Constructor | Description |
|-------------|-------------|
| `KeyboardOptions() TKeyboardOptions` | Creates options without any hints. |

The keyboard types are `KeyboardDefault`, `KeyboardAscii`, `KeyboardInteger`, `KeyboardFloat`, `KeyboardEMail`,
`KeyboardPhone`, `KeyboardSearch` and `KeyboardURL`.

## Methods

| Method | Description |
|--------|-------------|
| `AutoCorrectEnabled(autoCorrectEnabled bool) TKeyboardOptions` | Enables or disables auto-correction. |
| `Capitalization(capitalization bool) TKeyboardOptions` | Enables or disables automatic capitalization. |
| `KeyboardType(keyboardType KeyboardType) TKeyboardOptions` | Sets a hint for the keyboard type to show. |

## Related

- [Text Field](../../basic/text_field/)
- Tutorials: [Text field](/docs/examples/tutorial-12-textfield/)
