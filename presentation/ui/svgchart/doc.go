// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package svgchart renders bar, line and pie charts as static SVG images on the server.
//
// The charts take the same [chart.Chart] configuration and [chart.Series] data as the
// packages barchart, linechart and piechart, but do not depend on a charting library in
// the frontend: the Render method produces a plain SVG which is embedded through
// [ui.Image]. The SVG becomes part of the page, so theme colors like [ui.M0] are
// resolved by the browser and all text and lines follow the current text color, which
// makes the charts work in light and dark mode without any extra configuration.
//
// Because the result is a static image, there is no toolbar, no animation and no
// interactive tooltip. Bars, points and slices carry a title element, so browsers show
// the value on hover. The SVG() methods return the markup, e.g. to embed it in a
// PDF or a mail.
//
// Differences to the Apex based chart components:
//   - [chart.Chart.Downloadable] is ignored, there is no toolbar.
//   - A line chart has no concept of null data points, so there is no
//     ShowNullDataPoints option.
//   - A pie chart only renders its first series.
package svgchart
