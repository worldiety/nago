---
# Content is auto generated
# Manual changes will be overwritten!
title: Entry
---
Represents a single row or list item with optional headline,
supporting text/view, leading & trailing views, and an action handler.

## Constructors
### Entry
Entry creates a new full-width entry with default frame.

---
## Methods
| Method | Description |
|--------| ------------|
| `Action(fn func())` | Action sets a click/tap action handler. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame for the entry. |
| `Headline(s string)` | Headline sets the main title text of the entry. |
| `HeadlineView(view core.View)` |  |
| `Leading(v core.View)` | Leading sets an optional leading view (e.g. icon/avatar). |
| `SupportingText(s string)` | SupportingText sets an optional supporting text below the headline. |
| `SupportingView(view core.View)` | SupportingView sets an optional supporting view below the headline. |
| `Trailing(v core.View)` | Trailing sets an optional trailing view (e.g. button/chevron). |
---

