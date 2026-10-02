---
title: Installation
weight: 1
---

## Requirements

- **Go 1.27 or newer.** The `go` directive in Nago's
  [go.mod](https://github.com/worldiety/nago/blob/main/go.mod) requires it. With an older but recent Go
  installation, the Go command downloads the required toolchain automatically, unless you have set
  `GOTOOLCHAIN=local`.
- A current web browser.

Nothing else is needed: Nago contains no cgo code, brings its frontend as compiled assets and stores its data in
local files.

## Add Nago to your module

Create a module for your application and add Nago as a dependency:

```bash
mkdir myapp && cd myapp
go mod init example.com/myapp
go get go.wdy.de/nago@latest
```

`go.wdy.de/nago` is a vanity import path. It resolves to the public repository
[github.com/worldiety/nago](https://github.com/worldiety/nago), so you need no credentials or `GOPRIVATE`
settings.

The first build compiles a lot of packages and takes a while. Later builds are fast thanks to the Go build cache.

## Try an example without a project

Every tutorial in [`example/cmd`](https://github.com/worldiety/nago/tree/main/example/cmd) is a runnable
program. You can start one directly:

```bash
go run go.wdy.de/nago/example/cmd/tutorial-01-helloworld@latest
```

Then open [http://localhost:3000](http://localhost:3000).

## License

Nago is not open source software. Its [license](https://github.com/worldiety/nago/blob/main/LICENSE) (German)
says in short:

- Pupils, students and trainees may use Nago free of charge for their education, and schools, universities and
  other educational institutions may use it free of charge for teaching and learning.
- Everybody may **develop** applications with Nago free of charge.
- **Operating** an application built with Nago requires a written license agreement between the owner or
  operator of the software and worldiety.
- Sublicensing is not allowed, and all copyright and license notices must be kept.

{{< callout type="warning" >}}
This summary is not legal advice. The license text in the repository is binding. Contact
[worldiety](https://www.worldiety.de) before you put a Nago application into operation.
{{< /callout >}}

Third-party works contained in Nago, e.g. icons, remain under their own licenses.
