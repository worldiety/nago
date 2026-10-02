---
title: Workflow
weight: 3
---

## Install

The `speclink` command is installed like any Go tool:

```bash
go install github.com/worldiety/speclink/cmd/speclink@latest
```

Your project only depends on the declaration package, which `go mod tidy` adds once you import it:

```bash
go get github.com/worldiety/speclink/spec
```

The two are released together. The tool is deliberately not a dependency of your module: it is a compiler
frontend with its own dependencies, which have no place in the module graph of an application.

## Start a project

`speclink init` writes starting points for some profiles, but `go_nago_ddd1` has no template;
`speclink init -describe` lists what exists. For a Nago project, you set it up by hand:

1. Lay out the module as described in [Project layout](../../architecture/project-layout/).
2. Create `speclink.json` in the module root:

   ```json
   { "profile": "go_nago_ddd1" }
   ```

   speclink never guesses the profile. Without it, every command stops and lists the available ones.
3. Put the source documents into `requirements/_sources/`, write the requirements below `requirements/` and the
   annotation files next to your use cases, see [What you write](../files/).
4. Run the loop below until `verify` reports zero findings.

The first `verify` of a new project reports findings even if everything is written correctly: no test has been
recorded and no shape has been frozen yet. Clearing them is the first round of the loop.

## The build order

The order is fixed:

```bash
go build ./...
speclink verify ./...
go test -json ./... | speclink evidence
```

1. **`go build`** first. speclink needs code which compiles; if the build is broken, it refuses to run and says so.
2. **`speclink verify`** checks the requirement tree, the annotations, the architecture and the stored shapes. It
   exits with `1` on any finding and `0` otherwise.
3. **`speclink evidence`** reads the test output and records which passing tests demonstrated which requirements,
   in `speclink.lock`. speclink does not run the tests itself; your build or CI hands the results over.

With `-coverprofile`, `evidence` also records how much of each use case the tests executed:

```bash
go test -json -coverprofile=cover.out ./... | speclink evidence -coverprofile cover.out
```

`verify` ends with a summary. This is the one for tutorial-113 as it is today:

```
0 source segments (100% accounted), 7 constructs (100% bound), 6 normative requirements (100% covered, 0% verified, 0% demonstrated), 0 routes, 3 bindings, 24 findings
```

Every construct names a requirement and every requirement is implemented, but no test claims a requirement
(*verified*) and no passing test run has been handed to `speclink evidence` (*demonstrated*). See
[Why speclink](../why/#traceability-in-both-directions) for what the figures mean.

## Reading a finding

Each finding names a file and position, a rule and three parts: what is wrong, why it matters, and how to fix it.

```
path/to/file.go:12:6: [SPEC-V6-056] use case SubmitQuote has no permission of its own.
    A permission per use case is what makes authorisation assignable and auditable. …
    Add `PermSubmitQuote = permission.Declare[SubmitQuote]("…", name, description)` and check it in NewSubmitQuote.
```

The code `SPEC-<phase>-<number>` tells you the phase: `V1` the grammar of annotation files, `V3` a binding to an
illegal target, `V4` an annotation stating something already known, `V5` the requirement tree, `V6` the
specification and architecture rules. A later phase only runs once the earlier ones are clean. The rule name, e.g.
`K5-UC-PERMISSION`, is what you pass to `spec.Waive` when a rule genuinely cannot hold. `-format json` returns the
same findings for tools and agents.

## Recording changes: `freeze`

Some facts cannot be read from the current source: what a stored field used to be, what a requirement used to
say. `speclink freeze` records them in `speclink.lock`:

```bash
speclink freeze -n ./...                 # show what would be recorded
speclink freeze ./...                    # record it
speclink freeze -reviewer "Frau Meier" ./...   # and record that this person read the requirements
```

Run it after you committed to a stored shape, and after you re-read code whose requirement or source section was
reworded. Commit the lock file; its diff is the review.

## Other commands

| Command                              | Use it to                                                                 |
|--------------------------------------|---------------------------------------------------------------------------|
| `speclink requirements ./requirements/...` | check the requirement tree alone, while the code is not written yet |
| `speclink inventory ./...`           | list what speclink recognised, e.g. all use cases and whether they are bound |
| `speclink impact R-LIB-LEND`         | see what a requirement, a source section (`doc.md#anchor`) or a file reaches |
| `speclink generate -out SPECIFICATION.md ./...` | derive the specification document                          |
| `speclink diagrams ./...`            | write PlantUML sources of the context, building blocks and processes      |
| `speclink attest -origin llm ./app/library/...` | record that code was machine written                       |
| `speclink attest -reviewer "TS" LendBook` | record that a person has read a declaration                          |

All commands except `init` accept `-root` (default `.`) and `-config`, which reads the configuration from another file, so you
can try speclink on a project without changing it. `speclink <command> -h` lists the flags of a command.

## A PDF for review

`generate` writes Markdown by default. For a document to hand over, it writes Typst, which you compile into a PDF;
diagrams come from PlantUML. speclink runs neither tool itself, so you need PlantUML and Typst installed:

```bash
speclink diagrams -out build/puml -title "Library" ./...
plantuml -tsvg build/puml/*.puml

speclink generate -format typst -title "Library" -author "worldiety GmbH" -date 2026-10-02 \
  -figures build/puml -out build/spec.typ ./...
typst compile build/spec.typ build/spec.pdf
```

The date is passed in rather than read from the clock, so the same tree always produces the same document.

## Bringing in an existing project

A project with a lot of existing code is brought in package by package. Name the measured packages with `scope`
in `speclink.json`, or on the command line:

```bash
speclink verify ./app/library/... ./requirements/...
```

Both are the same: the whole module is loaded, only the named packages are measured, and the summary says how
many packages were left out. Widen the scope until it covers everything.
