---
title: Image Management
---

Image Management stores uploaded images and delivers them in a fitting resolution. An upload is validated and
stored as a source set: the original plus downscaled variants. Identical image data is stored only once.
Forms use it for fields of the type `image.ID`, for example avatars and logos.

Image Management has no admin UI.

## Enable

```go
images := std.Must(cfg.ImageManagement()) // application.ImageManagement
```

Image Management is always enabled when the application starts. `application.ImageManagement` has the field
`UseCases image.UseCases`.

## Use cases

| Use case       | Description                                                                                 |
|----------------|---------------------------------------------------------------------------------------------|
| `CreateSrcSet` | Validates an uploaded image and stores it with downscaled variants. `image.Options` limits the file size (default 32 MiB) and the resolution (default 3840 pixels). |
| `LoadSrcSet`   | Loads the source set of an image.                                                           |
| `LoadBestFit`  | Picks the variant which fits a given size and object fit best.                              |
| `OpenReader`   | Opens the raw image data.                                                                   |

## HTTP endpoint

Images are delivered by `/api/nago/v1/image` with the query parameters `src` (the image ID), `fit`, `w` and
`h`. Build the URL with `httpimage.URI(id, fit, w, h)` from `go.wdy.de/nago/application/image/http`.

{{< callout type="warning" >}}
Image Management declares no permissions and the endpoint does not check any. Everyone who knows an image ID
can load the image.
{{< /callout >}}

## Related

- [Tutorial: images](/docs/examples/tutorial-40-images/)
- [Tutorial: auto forms](/docs/examples/tutorial-48-form-auto/) uses `image.ID` fields with the styles
  `avatar` and `icon`.
