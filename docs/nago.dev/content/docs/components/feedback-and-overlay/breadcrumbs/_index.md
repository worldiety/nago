---
# Content is auto generated
# Manual changes will be overwritten!
title: Breadcrumbs
---
It displays a horizontal trail of items representing the user's navigation path
within the application. Each item is typically a link or label, separated by a
configurable gap, and the layout can be styled with frame and padding options.

## Constructors
### Breadcrumbs
Breadcrumbs creates a new breadcrumb trail with the given items.

---
## Methods
| Method | Description |
|--------| ------------|
| `ClampLeading()` | ClampLeading ensures that if the first entry is a default Item its title will be aligned to the leading of this component so that you can align the optical flight of text. |
| `Frame(frame ui.Frame)` | Frame defines the frame layout (size and positioning) of the breadcrumbs. |
| `Gap(l ui.Length)` | Gap sets the spacing between breadcrumb items. |
| `Item(title string, action func())` | Item appends a default button with the given title and text. Currently, this just defaults to a tertiary styled button. |
| `Padding(padding ui.Padding)` | Padding sets the inner padding around the breadcrumb trail. |
---

