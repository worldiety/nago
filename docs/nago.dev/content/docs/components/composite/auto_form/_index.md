---
# Content is auto generated
# Manual changes will be overwritten!
title: Auto Form
---


## Constructors
### Auto
Auto is similar to [crud.AutoBinding], however it does much less and just creates a form using
reflection from the given type. It does not require or understand entities and identities.
Also note, that the concrete type is inspected at runtime and not the given template T, which
is only needed for your convenience and to satisfy any concrete state type. Internally, everything gets evaluated
as [any]. T maybe also be an interface, thus ensure, that the state contains not a nil interface.

The current implementation only supports:
  - string fields
  - integer fields (literally)

Other features, which are supported by [crud.Auto] are not (yet) supported.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` |  |
| `Border(border ui.Border)` |  |
| `CardPadding(padding ui.Padding)` |  |
| `Frame(frame ui.Frame)` |  |
| `FullWidth()` |  |
| `Padding(padding ui.Padding)` |  |
| `Visible(visible bool)` |  |
| `WithFrame(fn func(ui.Frame) ui.Frame)` |  |
---

