---
title: Examples
weight: 5
sidebar:
  open: false
---

Each example is a small, runnable Nago application from
[`example/cmd`](https://github.com/worldiety/nago/tree/main/example/cmd). Its page shows what it does and its
complete source code.

## Run an example

You need a recent Go toolchain. Run any example directly from the module proxy, without a checkout:

```bash
go run go.wdy.de/nago/example/cmd/tutorial-01-helloworld@latest
```

Or, from a checkout of the [repository](https://github.com/worldiety/nago):

```bash
go run ./example/cmd/tutorial-01-helloworld
```

Then open [http://localhost:3000](http://localhost:3000). Set the environment variable `PORT` to use another
port. An example keeps its data in `~/.nago/<application ID>`, e.g. `~/.nago/de.worldiety.tutorial`, so it
survives a restart. Many examples share the same application ID and therefore the same data. Delete the
directory to start from scratch.

Some examples enable the standard systems and a bootstrap admin. Sign in as `admin@localhost` with the password
given to `EnableBootstrapAdmin` in the example's `main.go`.

## All examples

{{< example-gallery >}}
