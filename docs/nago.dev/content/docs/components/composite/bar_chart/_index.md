---
# Content is auto generated
# Manual changes will be overwritten!
title: Bar Chart
---
It represents categorical data with rectangular bars, supporting
horizontal or vertical orientation, stacked bars, and markers. The chart can be customized by providing multiple data series
and additional visual indicators (markers).

## Constructors
### BarChart
BarChart creates a new bar chart with default (vertical, non-stacked) configuration.

---
## Methods
| Method | Description |
|--------| ------------|
| `Chart(chart chart.Chart)` | Chart sets the underlying chart configuration for the bar chart. |
| `Horizontal(horizontal bool)` | Horizontal sets whether the bar chart is rendered horizontally. |
| `Markers(markers []Marker)` | Markers adds markers to the bar chart to highlight values or ranges. |
| `Series(series []chart.Series)` | Series defines the data series to be displayed in the bar chart. |
| `Stacked(stacked bool)` | Stacked sets whether multiple series are stacked instead of grouped. |
---

