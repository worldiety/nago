---
title: Getting Started
weight: 1
---

Nago is a Go framework by [worldiety](https://www.worldiety.de) for building business applications. You write
the whole application in Go, including its user interface. There is no JavaScript, HTML or CSS to maintain and no
REST API between your frontend and your backend.

## How it works

- **Server-driven UI.** Your Go code describes the user interface as a tree of views, e.g.
  `VStack(Text("hello"), PrimaryButton(save).Title("Save"))`. Nago renders this tree on the server and sends it
  over a websocket to the browser.
- **Embedded frontend.** A generic Vue.js frontend, shipped in the Nago module and embedded into your binary,
  displays the tree and sends user events like clicks and inputs back to the server. Your event handlers are
  plain Go functions.
- **Single binary.** `go build` produces one executable which contains the web server, the frontend and your
  application. Data is stored in a local directory, so you need no external database to get started.
- **Built-in systems.** Nago ships ready-to-use systems with their own admin user interface, e.g. user, role and
  permission management, sessions and login, mail, secrets, backups, templates and themes. You enable them with
  a single call on the configurator.

## Who it is for

Nago fits teams which build line-of-business applications, internal tools, admin portals or data-centric web
apps and prefer to stay in one language and one code base. You need basic Go knowledge; if you are new to Go,
take the [Go Tour](https://go.dev/tour/) first.

Nago is not meant for public websites optimized for search engines or for applications which need a custom,
pixel-perfect frontend written by frontend developers.

## Next steps

{{< cards >}}
  {{< card link="installation" title="Installation" icon="download" subtitle="Requirements, module and license." >}}
  {{< card link="first-app" title="Your first app" icon="play" subtitle="Run hello world and understand its parts." >}}
  {{< card link="../concepts" title="Concepts" icon="light-bulb" subtitle="How a Nago application works." >}}
{{< /cards >}}
