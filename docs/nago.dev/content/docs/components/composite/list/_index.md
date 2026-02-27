---
# Content is auto generated
# Manual changes will be overwritten!
title: List
---
It displays a vertical collection of rows, optionally with a caption and footer. A click handler can be attached to individual entries.

## Constructors
### List
List creates a new TList with the given entries as rows.

---
## Methods
| Method | Description |
|--------| ------------|
| `Caption(s core.View)` | Caption sets an optional caption view above the list. |
| `ColorBody(color ui.Color)` |  |
| `ColorCaption(color ui.Color)` |  |
| `ColorFooter(color ui.Color)` |  |
| `Footer(s core.View)` | Footer sets an optional footer view below the list. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the list. |
| `FullWidth()` | FullWidth expands the list to use the full available width. |
| `OnEntryClicked(fn func(idx int))` | OnEntryClicked sets a callback for when a row is clicked. |
| `With(fn func(c TList) TList)` |  |
---

