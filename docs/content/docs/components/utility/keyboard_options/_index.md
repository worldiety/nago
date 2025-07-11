---
# Content is auto generated
# Manual changes will be overwritten!
title: Keyboard Options
---
Keyboard Options defines configuration options for virtual keyboard behavior. It allows customization of capitalization, auto-correction, and keyboard type hints. These options are primarily used in text input components to enhance user experience.

## Constructors
### KeyboardOptions
```go
	KeyboardOptions()
```

---
## Methods
| Method | Description |
|--------| ------------|
| `AutoCorrectEnabled(autoCorrectEnabled bool)` | AutoCorrectEnabled enables or disables auto-correction. |
| `Capitalization(capitalization bool)` | Capitalization enables or disables automatic capitalization. |
| `KeyboardType(keyboardType KeyboardType)` | KeyboardType is a hint to the frontend. Technically, it is impossible to actually guarantee anything, and you have always to considers bugs and hacks:   - a malicious user may send you anything, which would otherwise not be possible (e.g. text instead of numbers)   - Android IME hints or keyboard types are never guaranteed. A user may install third-party keyboards which just ignore anything   - a user may inject anything using wrong autocompletion or the clipboard |
---
## Related

- [Keyboard Options](../../utility/keyboard_options/)
