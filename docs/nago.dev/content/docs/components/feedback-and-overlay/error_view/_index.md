---
# Content is auto generated
# Manual changes will be overwritten!
title: Error View
---
It shows an error message inside a styled container with optional
padding, spacing, borders, and layout configuration. Typically used
to surface application or runtime errors to the user.

## Constructors
### ErrorView
ErrorView returns a view which is suited to be displayed instead of your actual view in case of an unexpected
error. It is similar to the combined tuple of collecting errors using RequestSupport and showing them
through SupportRequestDialog. However, it returns an empty view, if err is nil. It returns a special view
when the permission
is denied and a support view in case of anything else to avoid leaking confidential error details.
Note, that unlike RequestSupport, each call to SupportView will immediately allocate a new SupportView, thus
better don't use it in loops to create error views over and over again.

---
