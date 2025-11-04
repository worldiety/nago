---
# Content is auto generated
# Manual changes will be overwritten!
title: Tabs
---
It manages the layout and navigation between different TPage elements,
including alignment, positioning, and spacing between the tab bar and content. An optional state can track the currently active tab index.

## Constructors
### Tabs
Tabs creates a new tab container with the given pages,
defaulting to leading alignment and a standard page-to-tab spacer.

---
## Methods
| Method | Description |
|--------| ------------|
| `ButtonAlignment(tabAlignment ui.Alignment)` | ButtonAlignment sets the alignment of the tab buttons within the button bar. Defaults to Leading. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the tabs container, including size and spacing. |
| `FullWidth()` | FullWidth sets the tabs container to span the full available width. |
| `InputValue(activeIdx *core.State[int])` | InputValue binds the tab container to an external state that tracks the index of the currently active page. |
| `PageTabSpace(space ui.Length)` | PageTabSpace is the amount of space between the tab button bar and the actual page content. Default to L32. Set to the empty string to disable any space. |
| `Position(pos ui.Position)` | Position sets the position of the tab bar (e.g., top, bottom, start, end). |
---

