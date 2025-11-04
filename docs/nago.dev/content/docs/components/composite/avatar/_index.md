---
# Content is auto generated
# Manual changes will be overwritten!
title: Avatar
---
This component displays a user or entity representation,
either as an image (via URL, raw data, or image ID) or as
a fallback with initials (paraphe). It can be styled with
frame, border, text size, and color, and also supports an
optional click action.

## Constructors
### Embed
Embed creates an avatar directly from raw image data.

### Text
Text creates a text-based avatar using initials derived from the given string.

### TextOrImage
TextOrImage creates an avatar from either an image (if provided) or falls back to a text-based avatar.

### URI
URI creates an avatar from a given image URL.

---
## Methods
| Method | Description |
|--------| ------------|
| `Action(fn func())` | Action sets an optional click action for the avatar. |
| `Border(border ui.Border)` | Border sets the border style of the avatar. |
| `Size(widthAndHeight ui.Length)` | Size sets the avatar's size and adjusts text size and image resolution accordingly. |
| `Style(style Style)` | Style sets the avatar’s border style (circle by default, rounded when specified). |
---

