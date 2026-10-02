---
title: PDF
---

A PDF viewer shows a PDF document from a URL inside your view, using the PDF viewer of the browser. Give it a
frame with a fixed height, otherwise the viewer has no space.

```go
PDF("https://pdfobject.com/pdf/sample.pdf").
	Frame(Frame{Width: L880, Height: L880, MaxWidth: Full})
```

## Constructors

| Constructor | Description |
|---|---|
| `func PDF(src core.URI) TPDF` | PDF creates a new PDF viewer component with the given source URL. |

## Methods

| Method | Description |
|---|---|
| `Frame(frame Frame) TPDF` | Frame sets the viewer's frame for sizing purposes. |
| `Src(src core.URI) TPDF` | Src sets the source URL of the PDF viewer. |

## Related

- [WebView](../webview/)
- [Image](../image/)
- Tutorials: [PDF viewer](/docs/examples/tutorial-106-pdf-viewer/)
