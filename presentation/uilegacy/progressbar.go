package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ProgressBar struct {
	id             ora.Ptr
	max            Float
	value          Float
	showPercentage Bool
	properties     []core.Property
}

func NewProgressBar(with func(progressBar *ProgressBar)) *ProgressBar {
	p := &ProgressBar{
		id:             nextPtr(),
		max:            NewShared[float64]("max"),
		value:          NewShared[float64]("value"),
		showPercentage: NewShared[bool]("showPercentage"),
	}

	p.value.Set(-1)
	p.max.Set(-1)
	p.showPercentage.Set(true)

	p.properties = []core.Property{p.max, p.value, p.showPercentage}
	if with != nil {
		with(p)
	}
	return p
}

func (p *ProgressBar) ID() ora.Ptr {
	return p.id
}

func (p *ProgressBar) Max() Float {
	return p.max
}

func (p *ProgressBar) Value() Float {
	return p.value
}

func (p *ProgressBar) ShowPercentage() Bool {
	return p.showPercentage
}

func (p *ProgressBar) Properties(yield func(core.Property) bool) {
	for _, property := range p.properties {
		if !yield(property) {
			return
		}
	}
}

func (p *ProgressBar) Render() ora.Component {
	return p.render()
}

func (p *ProgressBar) render() ora.ProgressBar {
	return ora.ProgressBar{
		Ptr:            p.id,
		Type:           ora.ProgressBarT,
		Max:            p.max.render(),
		Value:          p.value.render(),
		ShowPercentage: p.showPercentage.render(),
	}
}
