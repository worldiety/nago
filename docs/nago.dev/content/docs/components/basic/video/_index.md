---
title: Video
---

A video plays a video file from a URI, e.g. a static resource of your application. It lives in package
`presentation/ui/video` and maps to the HTML video element, so the supported formats depend on the browser. To
embed a video platform like YouTube, use a [WebView](../webview/) instead.

```go
//go:embed palm.mp4
var palm application.StaticBytes

// in your configurator
palmUri := cfg.Resource(palm)

// in your view
video.Video(palmUri).
	AutoPlay(true).
	Loop(true).
	Controls(true)
```

## Constructors

| Constructor | Description |
|---|---|
| `func Video(src core.URI) TVideo` | Creates a video for the given source. |

## Methods

| Method | Description |
|---|---|
| `AutoPlay(autoplay bool) TVideo` | Starts the playback automatically; browsers usually require `Muted` for that. |
| `Controls(controls bool) TVideo` | Shows the playback controls of the browser. |
| `Frame(frame ui.Frame) TVideo` | Sets the layout frame of the video. |
| `Loop(loop bool) TVideo` | Restarts the video when it ends. |
| `Muted(muted bool) TVideo` | Mutes the audio. |
| `PlaysInline(plays bool) TVideo` | Plays the video inline instead of fullscreen on mobile devices. |
| `Poster(poster core.URI) TVideo` | Sets an image which is shown until the video plays. |
| `Src(src core.URI) TVideo` | Sets the source of the video. |

## Related

- [Image](../image/)
- [WebView](../webview/)
- Tutorials: [Video](/docs/examples/tutorial-72-video/)
