---
# Content is auto generated
# Manual changes will be overwritten!
title: Progress
---
It represents the completion state of a task or process as a filled bar. The style, color, and background can be customized, and the value is given
as a floating-point percentage between 0. 0 and 1. 0.

## Constructors
### LinearProgress
LinearProgress creates a horizontal progress bar with default accent color,
card footer background, full width, and standard height.

---
## Methods
| Method | Description |
|--------| ------------|
| `BackgroundColor(color ui.Color)` | BackgroundColor sets the background color of the unfilled portion of the bar. |
| `Color(color ui.Color)` | Color sets the foreground color of the progress indicator. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the progress bar, including size and spacing. |
| `FullWidth()` | FullWidth sets the progress bar to span the full available width. |
| `Progress(v float64)` | Progress must be between 0 and 1. Values are clamped. |
| `Style(style Style)` | Style sets the visual style of the progress bar (e.g., horizontal or circular). |
---

