---
title: "Event Sourcing: Decide and Evolve"
weight: 85
---

Event sourcing with an aggregate: a command decides which events happen, each event evolves the aggregate, and `cfgevs.NewHandler` ties both together.

![Event Sourcing: Decide and Evolve](screenshot.webp)

{{< example-code >}}
