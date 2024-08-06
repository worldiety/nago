package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Scaffold struct {
	id                  ora.Ptr
	navigationComponent *Shared[*NavigationComponent]
	body                *Shared[core.View]
	properties          []core.Property
}

func NewScaffold(with func(scaffold *Scaffold)) *Scaffold {
	s := &Scaffold{
		id:                  nextPtr(),
		navigationComponent: NewShared[*NavigationComponent]("navigationComponent"),
		body:                NewShared[core.View]("body"),
	}

	s.properties = []core.Property{s.navigationComponent, s.body}

	if with != nil {
		with(s)
	}

	return s
}

func (s *Scaffold) Body() *Shared[core.View] {
	return s.body
}

func (s *Scaffold) NavigationComponent() *Shared[*NavigationComponent] {
	return s.navigationComponent
}

func (s *Scaffold) ID() ora.Ptr {
	return s.id
}

func (s *Scaffold) Type() string {
	return "Scaffold"
}

func (s *Scaffold) Properties(yield func(core.Property) bool) {
	for _, property := range s.properties {
		if !yield(property) {
			return
		}
	}
}

func (s *Scaffold) Render() ora.Component {
	var navigationComponent ora.Property[ora.NavigationComponent]
	if s.navigationComponent.Get() != nil {
		navigationComponent.Ptr = s.navigationComponent.id
		navigationComponent.Value = s.navigationComponent.v.renderNavigationComponent()
	}

	return ora.Scaffold{
		Ptr:                 s.id,
		Type:                ora.ScaffoldT,
		Body:                renderComponentProp(s.body, s.body),
		NavigationComponent: navigationComponent,
	}
}
