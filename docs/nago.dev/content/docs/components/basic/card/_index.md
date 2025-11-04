---
# Content is auto generated
# Manual changes will be overwritten!
title: Card
---
It represents a structured UI element that can display content in three
sections: a title, a body, and an optional footer. Cards are typically used to group related information or actions together
in a visually distinct block, and they support custom styling, padding,
and layout adjustments via frame and title style.

## Constructors
### Card
Card creates a new card with a title and default padding.

---
## Methods
| Method | Description |
|--------| ------------|
| `Body(view core.View)` | Body defines the main content area of the card. |
| `Footer(view core.View)` | Footer adds a footer view below the card body, typically for actions or secondary info. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame (size and positioning) of the card. |
| `ID(id string)` | ID sets the components unique identifier. |
| `Padding(padding ui.Padding)` | Padding overrides the default padding of the card and marks it as custom. |
| `Style(style TitleStyle)` | Style sets the title style of the card (e.g., heading level or visual variant). |
---

