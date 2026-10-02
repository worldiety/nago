---
title: Single Image Picker
---

The single image picker uploads one image and shows a preview with a button to remove it. Without an
image, it shows only the upload button. Uploaded images are stored by the image management, and the state
receives the ID of the new image.

![Single Image Picker](single_image_picker.webp)

```go
// createSrcSet, loadSrcSet and loadBestFit are the image use cases, see the note below.
var (
    createSrcSet image.CreateSrcSet
    loadSrcSet   image.LoadSrcSet
    loadBestFit  image.LoadBestFit
)

func view(wnd core.Window) core.View {
    img := core.AutoState[image.ID](wnd)

    return VStack(
        Text("Cover image"),
        form.SingleImagePicker(wnd, createSrcSet, loadSrcSet, loadBestFit, "cover-image", img.Get(), img),
    ).Alignment(Leading).Gap(L8).Frame(Frame{Width: L400})
}
```

Pass the use cases of `cfg.ImageManagement()`, or `nil` to resolve them from the window context. The
latter requires that the image management is enabled, e.g. by `cfg.StandardSystems()`.

## Constructors

```go
func SingleImagePicker(wnd core.Window, setCreator image.CreateSrcSet, loadSrcSet image.LoadSrcSet, loadBestFit image.LoadBestFit, selfId string, id image.ID, state *core.State[image.ID]) TSingleImagePicker
```

SingleImagePicker creates a new TSingleImagePicker with the given configuration.

## Methods

| Method | Description |
|--------|-------------|
| `AccessibilityLabel(label string) ui.DecoredView` | AccessibilityLabel sets the accessibility label for the single image picker. |
| `Border(border ui.Border) ui.DecoredView` | Border sets the border styling of the single image picker. |
| `Frame(frame ui.Frame) ui.DecoredView` | Frame sets the frame of the single image picker directly. |
| `Padding(padding ui.Padding) ui.DecoredView` | Padding sets the padding of the single image picker. |
| `Visible(visible bool) ui.DecoredView` | Visible toggles the visibility of the single image picker. |
| `WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView` | WithFrame updates the frame of the single image picker using a transformation function. |

## Related

- [Avatar Picker](../avatar_picker/)
- [Image](../../basic/image/)
