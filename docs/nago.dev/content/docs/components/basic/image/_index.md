---
# Content is auto generated
# Manual changes will be overwritten!
title: Image
---
It displays raster or vector images with support for light/dark mode variants,
SVG embedding, object fit, custom colors, borders, padding, and accessibility. The component can also be toggled invisible when needed.

## Constructors
### Image
Image creates a new image with a default frame size of Auto x L160.

### ImageIcon
ImageIcon renders default with L24/L24 size and is invisible if svg is empty.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label for screen readers. See also https://www.w3.org/WAI/tutorials/images/decision-tree/. |
| `Adaptive(dark image.ID, light image.ID)` | Adaptive sets the image to use different sources for light and dark themes. It generates URIs for both light and dark variants with fixed dimensions (512x512) and no fitting, then applies them to the image. |
| `Border(border Border)` | Border sets the border styling of the image. |
| `Embed(buf []byte)` | Embed encodes the given buffer within the components attributes. This may be fine to load small images synchronously, but it may break the channel, the server or the frontend, if too large. Better use [application.Resource] for large static images. Embedding image data in the range of 100-200 byte is totally fine, though. The resource URI alone is already about 100 characters long. Usually, embedding SVGs in the range of 1-2KiB is also fine. To optimize render performance, the system uses a special caching technique. Important: due to caching, do not submit ever-changing SVGs, because the backend and the frontend may suffer from cache overflow. This will half the typical required bandwidth for icon heavy use cases. The larger the SVG, the better the effect. See also [TImage.EmbedAdaptive]. |
| `EmbedAdaptive(dark []byte, light []byte)` | EmbedAdaptive is like [TImage.Embed] but picks whatever fits best. |
| `FillColor(color Color)` | FillColor set the internal fill color value and is only applicable for embedded SVG images, which use fill=currentColor. Otherwise, a broken or no effect will be seen. |
| `Frame(frame Frame)` | Frame sets the layout frame of the image, including size and positioning. |
| `ObjectFit(fit ObjectFit)` | ObjectFit sets how the image should be resized or scaled inside its frame (e.g., contain, cover, or none). |
| `Padding(padding Padding)` | Padding sets the inner spacing around the image. |
| `StrokeColor(color Color)` | StrokeColor set the internal stroke color value and is only applicable for embedded SVG images, which use fill=strokeColor. Otherwise, a broken or no effect will be seen. |
| `URI(uri core.URI)` | URI can be used for static image resources which are not provided by the ui component itself. The source may be a hand written REST endpoint or even any third-party resource from a different domain. If you need optimized data access and caching policies, you have to use this way. See also [core.Window.AsURI] for an uncached dynamically delivered image resource. |
| `URIAdaptive(dark core.URI, light core.URI)` | URIAdaptive is like [TImage.Embed] but picks whatever fits best. |
| `Visible(b bool)` | Visible controls the visibility of the image; setting false hides it. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the image's frame and returns the updated component. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

## Tutorials
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
