// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package linechart

import (
	"go.wdy.de/nago/presentation/proto"
)

type Curve int

func (c Curve) ora() proto.LineChartCurve {
	return proto.LineChartCurve(c)
}

const (
	CurveStraight = Curve(proto.LineChartCurveStraight)
	CurveSmooth   = Curve(proto.LineChartCurveSmooth)
	CurveStepline = Curve(proto.LineChartCurveStepline)
)
