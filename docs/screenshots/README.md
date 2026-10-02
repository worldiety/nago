# nago.dev screenshots

Renders all screenshots of nago.dev reproducibly, so that they all share the same viewport, scale, language
and crop. Each example listed in `shots/*.yaml` is built, started with an empty data directory and captured
with a headless Chrome in English.

```bash
cd docs/screenshots
go run .                                     # all shots
go run . -only tutorial-11-buttons           # a single example
go run . -only example/gallery/basic         # a component gallery
go run . -f shots/systems.yaml               # a single manifest
go run . -only tutorial-11-buttons -headful  # watch the browser, useful to debug steps
```

Requires Go and a local Chrome or Chromium.

Each area of the site has its own manifests, e.g. `examples-*.yaml`, `components-*.yaml` and `systems.yaml`.

## Declaring a shot

```yaml
defaults:                               # optional, per manifest
  width: 1200

shots:
  - example: tutorial-11-buttons        # directory below example/cmd, or a package like example/gallery/basic
    out: content/docs/examples/tutorial-11-buttons/screenshot.webp  # relative to docs/nago.dev
    path: /                             # optional url path
    width: 1200                         # viewport in CSS pixels, scale 2 by default
    dark: true                          # emulate prefers-color-scheme: dark
    crop: auto                          # auto | viewport | page | <css selector>
    admin: true                         # enable the bootstrap admin admin@localhost
    steps:                              # optional interactions before the capture
      - login: true                     # sign in as admin@localhost
      - clickText: Open dialog
      - type: { selector: "input", text: "hello" }
      - wait: 300ms
```

`crop: auto` captures the bounding box of everything that is visible (text, media and boxes which differ
from the page background) plus `padding`. Open dialogs dim the whole page, so their shots cover the viewport.

The available steps are `login`, `click` (css selector), `clickText`, `hover`, `type`, `wait`, `js` and
`goto`. All shots of one example run in the same browser tab, one after another, so a `login` in the first
shot is valid for the following ones.

## Component gallery

`example/gallery/<group>` (basic, layout, composite) are small apps with one route per component, e.g.
`/button`. A component registers its demo with `app.Register` in its own file. The galleries exist only to
render the component screenshots and are not listed as tutorials.
