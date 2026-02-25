---
# Content is auto generated
# Manual changes will be overwritten!
title: Auto Form
---
This component renders a form for type T driven by reflection,
bound to a state and configurable via AutoOptions.

## Constructors
### Auto
Auto is similar to [crud.AutoBinding], however it does much less and just creates a form using
reflection from the given type. It does not require or understand entities and identities.
Also note, that the concrete type is inspected at runtime and not the given template T, which
is only needed for your convenience and to satisfy any concrete state type. Internally, everything gets evaluated
as [any]. T maybe also be an interface, thus ensure, that the state contains not a nil interface.

The current default implementation only supports:
  - string fields
  - integer fields (literally)
  - string slices
  - bool fields
  - float fields

Other features, which are supported by [crud.Auto] are not (yet) supported.

Supported field tags:
  - visible:"true"|"false" defaults to true
  - section:"some text" defaults to zero
  - label:"string literal"|"i18n key" defaults to Field name
  - source:"source id" defaults to zero, only applicable to fields with underlying type string or []string. The
    source must be provided using [Configuration.AddContextValue] as type [AnyUseCaseList].
  - lines:"integer" only applicable to fields with underlying type string or []string and defaults to zero which
    renders as a single line. 1 also renders as single line, but uses a multiline input element. Defaults to zero
    for string types and to 5 for []string types.
  - value:"string literal"|"bool literal"|"number literal" only applicable for fields with the according underlying
    type. Defaults to the zero value of the underlying type.
  - dialogOptions:"large|larger|xlarge|xxlarge" is only supported for source picker.

The actual support may vary and depends on [AutoOptions.Renderers].

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the accessibility label for the auto form. |
| `Border(border ui.Border)` | Border sets the border styling of the auto form. |
| `CardPadding(padding ui.Padding)` |  |
| `Frame(frame ui.Frame)` | Frame sets the frame of the auto form directly. |
| `FullWidth()` |  |
| `Padding(padding ui.Padding)` | Padding sets the padding of the auto form. |
| `Visible(visible bool)` | Visible toggles the visibility of the auto form. |
| `WithFrame(fn func(ui.Frame) ui.Frame)` | WithFrame updates the frame of the auto form using a transformation function. |
---

