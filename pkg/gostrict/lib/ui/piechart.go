// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/ui/piechart"
)

// TPieChart is a pie chart, which shows the shares of a series as slices, see piechart.TPieChart of nago.
type TPieChart struct{ ref js.Ref }

func (TPieChart) isView() {}

// ShowAsDonut shows the chart as a donut, which has a hole in the middle.
//
//gostrict:host
func (c TPieChart) ShowAsDonut(showAsDonut bool) TPieChart {
	return c.with(func(v piechart.TPieChart) piechart.TPieChart { return v.ShowAsDonut(showAsDonut) })
}

// ShowDataLabels shows the labels of the slices.
//
//gostrict:host
func (c TPieChart) ShowDataLabels(showDataLabels bool) TPieChart {
	return c.with(func(v piechart.TPieChart) piechart.TPieChart { return v.ShowDataLabels(showDataLabels) })
}

// ShowAbsoluteValues shows the absolute values instead of the percentages.
//
//gostrict:host
func (c TPieChart) ShowAbsoluteValues(showAbsoluteValues bool) TPieChart {
	return c.with(func(v piechart.TPieChart) piechart.TPieChart { return v.ShowAbsoluteValues(showAbsoluteValues) })
}
