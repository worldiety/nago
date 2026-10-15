// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

// ChartRoundingType tells, how the labels of a chart are rounded, like chart.RoundingType of nago.
type ChartRoundingType int32

const (
	ChartRoundingNone    ChartRoundingType = 0
	ChartRoundingRound   ChartRoundingType = 1
	ChartRoundingFloor   ChartRoundingType = 2
	ChartRoundingCeiling ChartRoundingType = 3
)
