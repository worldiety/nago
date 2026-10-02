---
title: Store Inspector Benchmark
weight: 105
---

Fills an entity store with 4 million entries in the background, to measure the store inspector of the admin
center with a huge data set. The inspector is expected to stay fast, because it only iterates ids and pages the
result.

1. Start the example. On the very first start it populates the `BenchRecord` store in the background; the log
   shows the progress and the total duration.
2. Sign in as `admin@localhost` with the bootstrap password from `main.go`. The bootstrap admin receives all
   `nago.*` permissions, including the one for the inspector.
3. Open the admin center, go to the store inspector and select the `BenchRecord` store.

The keys are zero-padded (`record-000000000000`), so the lexicographic order of the store equals the numeric
order, which makes paging easy to follow and reproducible across runs.

{{< example-code >}}
