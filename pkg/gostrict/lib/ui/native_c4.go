// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/presentation/ui/slider"
	"go.wdy.de/nago/presentation/ui/stepper"
	"go.wdy.de/nago/presentation/ui/switcher"
	"go.wdy.de/nago/presentation/ui/tabs"
)

// The conversions of batch c4, see native.go.

func (v SliderRangeSliderValue) nago() slider.RangeSliderValue {
	return slider.RangeSliderValue{From: v.From, To: v.To}
}

func sliderRangeSliderValueFrom(v slider.RangeSliderValue) SliderRangeSliderValue {
	return SliderRangeSliderValue{From: v.From, To: v.To}
}

func stepperStepsOf(steps []TStepperStep) []stepper.TStep {
	res := make([]stepper.TStep, 0, len(steps))
	for _, s := range steps {
		res = append(res, get[stepper.TStep](s.ref))
	}
	return res
}

func switcherPagesOf(pages []TSwitcherPage) []switcher.TSwitcherPage {
	res := make([]switcher.TSwitcherPage, 0, len(pages))
	for _, p := range pages {
		res = append(res, get[switcher.TSwitcherPage](p.ref))
	}
	return res
}

func tabsPagesOf(pages []TTabsPage) []tabs.TPage {
	res := make([]tabs.TPage, 0, len(pages))
	for _, p := range pages {
		res = append(res, get[tabs.TPage](p.ref))
	}
	return res
}
