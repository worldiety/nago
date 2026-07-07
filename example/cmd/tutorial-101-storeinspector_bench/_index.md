---
# Static and dynamic content mixed!
# Use the shortcode {{< include-code "main.go" >}} to include the content of the file as a go-code block.
title: Tutorial 101 - Store Inspector Benchmark
---

This example populates a single entity store with a large amount of entries
(4 million by default) so that the performance of the store inspector
(Admin Center -> Inspektor -> Stores) can be investigated with a realistic,
huge dataset.

The store inspector is expected to be efficient, because it iterates ids only
and pages the result. This playground makes it easy to reproduce and measure the
actual runtime behavior with millions of entries.

## How to run

1. Start the example. On the very first start it populates the `BenchRecord`
   store in the background. Watch the log for the progress and the total
   duration.
2. Log in as `admin@localhost` with the password `%6UbRsCuM8N$auy`. The
   bootstrap admin automatically receives all `nago.*` permissions, including
   `nago.data.inspector`.
3. Open the Admin Center, go to `Inspektor -> Stores` and select the
   `BenchRecord` store to inspect the paging and selection behavior.

The keys are zero-padded (`record-000000000000`), so the natural lexicographic
order of the store equals the numeric order, which makes paging easy to follow
and fully reproducible across runs.

## Example
{{< include-code "main.go" >}}
