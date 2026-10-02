---
title: Redraw
---

`RedrawAtFixedRate` renders the current view again and again at the given rate, as long as the view is shown. Use
it for views which display a value that changes over time without a state change, like a clock or live
measurements. It returns the given view unchanged.

```go
func clock(wnd core.Window) core.View {
	return RedrawAtFixedRate(wnd, time.Second,
		Text(time.Now().Format(time.TimeOnly)),
	)
}
```

## Functions

| Function | Description |
|----------|-------------|
| `RedrawAtFixedRate[T core.View](wnd core.Window, rate time.Duration, v T) T` | Passes the given view and causes a redraw at the given rate. |

A rate faster than the frame rate of the application has no visible effect. Each redraw renders the whole view,
so keep the rate as low as possible.

## Related

- [Conditional rendering](../conditionals/), [Countdown](../../composite/countdown/)
- Tutorials: [Redraw](/docs/examples/tutorial-24-redraw/)
