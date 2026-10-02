# Maintaining nago.dev

The documentation on [nago.dev](https://www.nago.dev) is written and maintained by hand. Nothing is generated
from the source code, except that the tutorial sources are shown as they are in `example/cmd`. Once per quarter
the commits since the last review are checked and the docs are updated where necessary.

## Layout

| Path                           | Purpose                                                              |
|--------------------------------|----------------------------------------------------------------------|
| `docs/nago.dev/`               | Hugo site with the [Hextra](https://imfing.github.io/hextra/) theme  |
| `docs/nago.dev/content/docs/`  | the documentation, see sections below                                |
| `docs/screenshots/`            | renders all screenshots, see its README                              |
| `docs/serve.go`                | tiny server which embeds the built site for deployment on nago.app   |
| `example/cmd/`                 | the tutorials, mounted into Hugo and shown by `{{< example-code >}}` |
| `example/gallery/<group>/`     | one route per component, only used for the component screenshots    |
| `docs/.last-reviewed-commit`   | the commit up to which the docs have been reviewed                   |

Sections of `content/docs/`:

1. `getting-started/` what Nago is, installation, the first app.
2. `concepts/` how a Nago application works: configurator, views, state, navigation, use cases and
   permissions, persistence, theming, localization, deployment.
3. `components/` one page per UI component, grouped into basic, layout, composite and feedback-and-overlay.
4. `systems/` one page per system (user management, mail, ...).
5. `examples/` one page per tutorial in `example/cmd`.

## Build and preview

```bash
cd docs/nago.dev
hugo server            # preview on http://localhost:1313
hugo build --minify    # what CI does
```

CI (`.github/workflows/docs.yml`) builds and deploys on every push to `main` which touches the site, the
server or the examples. Use the Hugo version pinned there.

## Writing conventions

- English, second person, present tense. Short sentences, no marketing.
- Every statement must be true for the current source code. Verify names and signatures in the code before
  writing them down; never copy from older docs without checking.
- Code snippets must compile against the current API. Prefer showing a real tutorial with
  `{{< example-code "tutorial-xy" >}}` over a hand-written snippet.
- Use Hextra shortcodes (`callout`, `cards`, `tabs`, `steps`) where they help. CSS classes of the theme use
  the Tailwind v4 prefix `hx:`, e.g. `hx:mt-4`.
- Link between pages with relative links or `{{< ref >}}`. Link to code on
  `https://github.com/worldiety/nago/tree/main/...`.
- Front matter: `title`, optional `weight` and `linkTitle`. Do not set `prev`/`next`, the theme derives them.

### Component pages

`content/docs/components/<group>/<component>/_index.md`:

1. one paragraph: what it is and when to use it,
2. a screenshot rendered from `example/gallery/<group>` and the matching snippet,
3. **Constructors**: each exported constructor with its signature and a one-line description,
4. **Methods**: a table of all exported methods of the component type with the first sentence of its godoc,
5. **Related**: links to related components and tutorials.

The method tables are maintained by hand. When reviewing, compare them with the exported methods in
`presentation/ui/...`, e.g. `go doc ./presentation/ui.TButton` from the repository root.

### System pages

`content/docs/systems/<system>/_index.md`: what the system does, how to enable it on the configurator, its
use cases, the permissions it declares, and screenshots of its UI. Admin UIs are captured with `admin: true`
and a `login` step.

### Example pages

`content/docs/examples/<tutorial>/index.md` (the bundle name equals the directory in `example/cmd`):

```markdown
---
title: Buttons
---

One or two sentences what the tutorial shows.

![Buttons](screenshot.webp)

{{< example-code >}}
```

## Screenshots

All screenshots are rendered by `docs/screenshots`, never taken by hand, so that they share viewport, scale,
language and crop. Declare each image in `docs/screenshots/shots/<area>.yaml`, render it with
`go run . -only <example>` and commit the `.webp` next to the page that uses it. Delete images which are no
longer referenced.

## Quarterly review

1. `git log --oneline $(cat docs/.last-reviewed-commit)..HEAD` and read the commits which touch
   `application/`, `presentation/`, `example/` or the configurator.
2. For each user-visible change decide whether the docs are affected: new or renamed components, methods,
   systems, use cases, configuration, environment variables, new or removed tutorials.
3. Update the affected pages and method tables, add pages for new tutorials and components, remove pages of
   removed ones.
4. Re-render the screenshots of everything whose UI changed. When in doubt, re-render all: `go run .`.
5. `hugo build` must pass without errors or warnings.
6. Write the reviewed commit into `docs/.last-reviewed-commit` and commit everything together.
