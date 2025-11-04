---
# Content is auto generated
# Manual changes will be overwritten!
title: Countdown
---
It displays a timer counting down from a specified duration,
optionally showing days, hours, minutes, and seconds. The component
supports custom colors, styling, progress indicators, and an action
callback to be executed when the countdown completes.

## Constructors
### CountDown
CountDown creates a new countdown timer initialized with the given duration.
By default, days, hours, minutes, and seconds are all displayed.

---
## Methods
| Method | Description |
|--------| ------------|
| `Action(action func())` | Action sets the callback function to be executed when the countdown ends. |
| `Days(show bool)` | Days toggles whether the countdown displays days. |
| `Done(done bool)` | Done marks the countdown as finished, overriding its active state. |
| `Frame(frame Frame)` | Frame sets the layout frame of the countdown, including size and positioning. |
| `Hours(show bool)` | Hours toggles whether the countdown displays hours. |
| `Minutes(show bool)` | Minutes toggles whether the countdown displays minutes. |
| `ProgressBackground(background Color)` | ProgressBackground sets the background color of the countdown's progress indicator. |
| `ProgressColor(foreground Color)` | ProgressColor sets the foreground color of the countdown's progress indicator. |
| `Seconds(show bool)` | Seconds toggles whether the countdown displays seconds. |
| `SeparatorColor(color Color)` | SeparatorColor sets the color of separators (e.g., colons) in the countdown display. |
| `Style(style CountDownStyle)` | Style sets the visual style of the countdown (e.g., text-only or with progress). |
| `TextColor(color Color)` | TextColor sets the color of the countdown text. |
---

## Related
- [Frame](../../layout/frame/)

