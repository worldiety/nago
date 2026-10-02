---
title: speclink
weight: 4
---

[speclink](https://github.com/worldiety/speclink) is the tool worldiety uses to keep requirements, code, tests and
documentation of a Nago project in step. It runs after the Go compiler as part of the build and fails when they
disagree.

In one sentence: you write the requirements as Go values, you state next to the code which requirement it was
written for, and speclink checks that every requirement is implemented and tested, that every piece of code has a
reason, and that the [architecture](../architecture/) is followed. From the same source it derives the
specification document.

{{< cards >}}
  {{< card link="why" title="Why speclink" icon="question-mark-circle" subtitle="The problem it solves and how." >}}
  {{< card link="files" title="What you write" icon="document-text" subtitle="Requirements, annotations, sources and tests." >}}
  {{< card link="workflow" title="Workflow" icon="terminal" subtitle="Installing and running speclink in a project." >}}
  {{< card link="example" title="Example" icon="code" subtitle="A requirement traced through tutorial-113." >}}
{{< /cards >}}

## At a glance

- **One quality gate.** Every finding is an error. There are no warnings and no severities.
- **One architecture.** The profile `go_nago_ddd1` defines the [architecture](../architecture/) centrally, instead
  of each team writing its own conventions.
- **Traceability.** Each requirement is linked to the section of the document it came from, to the code which
  implements it and to the test run which demonstrated it.
- **Explainability.** For each use case you can ask why it exists. A Nago application can answer that at run time,
  see [Requirements](/docs/systems/speclink_management/).
- **Derived documents.** The specification for review is generated, as Markdown or as a PDF via Typst.
- **Impact and drift.** speclink shows what a change reaches and notices when a requirement or its source text
  changed under the code.

## speclink and Nago

Nago itself only depends on the small module `github.com/worldiety/speclink/spec`, which contains the declarations
(`spec.Declare`, `spec.For`, ...) and nothing but the standard library. The `speclink` command is a separate tool
which you install and run in your project; it is not part of your application's module graph.

speclink knows Nago's types: it recognises use cases, permissions, repositories, events and projections by their
Nago types, so you do not annotate them. It supports other profiles as well, but on nago.dev only `go_nago_ddd1`
is described.
