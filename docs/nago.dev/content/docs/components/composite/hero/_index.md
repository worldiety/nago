---
title: Hero
---

A hero is the large banner at the top of a landing page with a title, a subtitle, actions and an
optional image at the side or in the background.

![Hero](hero.webp)

```go
func view(wnd core.Window) core.View {
    return hero.Hero("Ship your next app in days").
        Subtitle("Nago renders your Go code as a modern web app. No JavaScript, no REST API, no build pipeline.").
        Actions(
            PrimaryButton(func() {}).Title("Get started").PostIcon(icons.ArrowRight),
            SecondaryButton(func() {}).Title("Read the docs"),
        ).
        SideSVG(icons.RocketLaunch).
        Alignment(Leading).
        Frame(Frame{Width: L1200, MaxWidth: Full})
}
```

## Constructors

```go
func Hero(title string) THero
```

Hero creates a new THero with the given title and a default full-width height of 320.

## Methods

| Method | Description |
|--------|-------------|
| `Actions(actions ...core.View) THero` | Actions sets the action buttons or links of the hero section. |
| `Alignment(alignment ui.Alignment) THero` | Alignment sets the position of the text block. |
| `BackgroundColor(color ui.Color) THero` | BackgroundColor sets the background color. |
| `BackgroundImage(img core.URI) THero` | BackgroundImage places a fit-cover image into the background. |
| `ForegroundColor(col ui.Color) THero` | ForegroundColor sets the color of the side image or SVG. |
| `ForegroundColorAdaptive(onLight, onDark ui.Color) THero` | ForegroundColorAdaptive sets the foreground color for light and dark mode. |
| `Frame(frame ui.Frame) THero` | Frame sets the frame of the hero section. |
| `Padding(padding ui.Padding) THero` | Padding sets the inner padding. |
| `SideImage(img ui.TImage) THero` | SideImage sets a side image for the hero section, which is displayed alongside the text content. |
| `SideSVG(svg core.SVG) THero` | SideSVG shows the given SVG next to the text. |
| `SideView(img core.View) THero` | SideView shows an arbitrary view next to the text. |
| `Subtitle(text string) THero` | Subtitle sets the subtitle text of the hero section. |
| `SubtitleView(view core.View) THero` | SubtitleView sets a custom view instead of the subtitle text. |
| `TextColor(color ui.Color) THero` | TextColor sets the color of title and subtitle. |

## Related

- [Image](../../basic/image/)
- Tutorial [tutorial-108-hero](/docs/examples/tutorial-108-hero/)
