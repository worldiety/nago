---
# Content is auto generated
# Manual changes will be overwritten!
title: Frame
---
Frame defines the sizing constraints and fixed dimensions for a UI element. It allows you to specify minimum and maximum width/height, as well as fixed
dimensions. Frames are used to control layout behavior and responsiveness. All fields are optional. If a field is zero, it will not constrain the layout.

## Methods
| Method | Description |
|--------| ------------|
| `FullHeight()` | FullHeight sets the frame's height to 100% of the available space. |
| `FullWidth()` | FullWidth sets the frame's width to 100% of the available space. |
| `IsZero()` | IsZero returns true if all fields of the Frame are unset (zero value). |
| `Large()` | Large sets the max width to 560dp (35rem) and Width to Full. |
| `Larger()` | Larger sets the width to 880dp (55rem) and Width to Full. |
| `MatchScreen()` | MatchScreen sets the frame to match the full viewport height and width. This is useful for fullscreen layouts or sections that should fill the screen. |
| `Size(w Length, h Length)` | Size sets both Width and Height to the given values and returns the updated Frame. |
---

## Tutorials
- [tutorial-01-helloworld](../../../examples/tutorial-01-helloworld)
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
- [tutorial-54-codeeditor](../../../examples/tutorial-54-codeeditor)
