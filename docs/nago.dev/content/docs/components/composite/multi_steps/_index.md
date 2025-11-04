---
# Content is auto generated
# Manual changes will be overwritten!
title: Multi Steps
---
This component manages and displays a sequence of steps,
tracking the active step index, available steps, and a completion button. It can also apply custom logic to determine if a step can be shown.

## Constructors
### MultiSteps
MultiSteps creates a new TMultiSteps with the provided steps.

---
## Methods
| Method | Description |
|--------| ------------|
| `ButtonDone(view core.View)` | ButtonDone sets the view to display when the steps are completed. |
| `CanShow(fn func(currentIdx intwantedIndex int) bool)` | CanShow sets a predicate to control whether a given step can be shown. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the multi-steps component. |
| `InputValue(idx *core.State[int])` | InputValue binds the active step index state to the multi-steps component. |
---

