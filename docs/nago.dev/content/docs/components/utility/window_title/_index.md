---
title: Window Title
---

A window title sets the title of the browser tab. It renders nothing visible, so place it anywhere in your view.
The headings `H1` and `Heading(1, ...)` set the window title automatically.

```go
return VStack(
	WindowTitle("Orders"),
	Text("All your orders"),
)
```

## Constructors

| Constructor | Description |
|-------------|-------------|
| `WindowTitle(title string) TWindowTitle` | Creates a new TWindowTitle with the given title text. |

## Methods

| Method | Description |
|--------|-------------|
| `Title(title string) TWindowTitle` | Changes the title text. |

## Related

- [Text](../../basic/text/)
- Tutorials: [Scaffold](/docs/examples/tutorial-17-scaffold/), [Hydration](/docs/examples/tutorial-91-hydration/)
