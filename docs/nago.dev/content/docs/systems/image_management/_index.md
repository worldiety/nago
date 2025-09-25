---
title: Image Management
galleryCodeUsageExample:
  - src: "/images/systems/image_management/galleries/code_usage/example.png"
---

The Image Management system provides backend functionality for handling image uploads, storage, and serving. It is automatically initialized at application startup and is always available.

## Functional areas
Session Management provides the following key functions:

### Image storage
- Store images in a blob store
- Manage SourceSets (image pyramids with multiple resolutions)
- Supports `png` and `jpeg` formats

### Image processing
- Automatically generate thumbnails / scaled versions
- Object-fit support for responsive layouts (`FitCover`, `FitNone`)
- Validate file size and dimensions

### Use cases
- `CreateSrcSet`: creates an SrcSet from an uploaded file
- `LoadSrcSet`: loads a previously stored SrcSet
- `LoadBestFit`: returns the best image candidate for given dimensions and object-fit
- `OpenReader`: reads raw image data from the blob store

### HTTP endpoint
The HTTP endpoint `/api/nago/v1/image` exposes stored images over HTTP.
- Accepts query parameters: `src` (image ID), `fit` (object-fit), `w` (width), `h` (height)
- Returns the best-fit image from a SrcSet
- Sets caching headers for long-term caching (1 year)

This endpoint allows frontend pages to render images dynamically without embedding raw image data.

## Code usage
Example: create a SrcSet from an embedded PNG and render it in a page.

```go
package main

import(
	_ "embed"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

//go:embed example.png
var embeddedImage []byte

func main() {
	application.Configure(func (cfg *application.Configurator) {
		cfg.SetApplicationID("image.management.example")
		
		file := core.MemFile{
			Filename:     "example.png",
			MimeTypeHint: "image/png",
			Bytes:        embeddedImage,
		}
		
		imgManagement := std.Must(cfg.ImageManagement())
		
		srcSet, err := imgManagement.UseCases.CreateSrcSet(user.SU(), image.Options{}, file)
		if err != nil {
			panic(err)
		}
		
		cfg.RootViewWithDecoration("image_example", func(wnd core.Window) core.View {
			imgURL := httpimage.URI(srcSet.ID, image.FitCover, 800, 600)
			return ui.VStack(
				ui.Text("Example Image"),
				ui.Image().URI(imgURL),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).
		Run()
}
```

{{< swiper name="galleryCodeUsageExample" loop="false" >}}

{{< callout type="info" >}}
Image Management runs entirely on the backend. There is no UI in the Admin Center. Images are served through the HTTP endpoint or programmatically via the UseCases API.
{{< /callout >}}

## Dependencies
**Requires:**
- No other systems

**Is required by:**
- [Signature Management](../signature_management/)

{{< callout type="info" >}}
Image Management is initialized automatically at application startup via Run().
{{< /callout >}}

## Activation
This system is activated via:
```go
std.Must(cfg.ImageManagement())
```

```go
ImageManagement := std.Must(cfg.ImageManagement())
```