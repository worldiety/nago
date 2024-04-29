package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Scaffold struct {
	id                  ora.Ptr
	navigationComponent EmbeddedNavigationComponent
	body                *Shared[core.Component]
	properties          []core.Property
}

func NewScaffold(with func(scaffold *Scaffold)) *Scaffold {
	s := &Scaffold{
		id:                  nextPtr(),
		navigationComponent: NewShared[ora.NavigationComponent]("navigationComponent"),
		body:                NewShared[core.Component]("body"),
	}

	s.properties = []core.Property{s.navigationComponent, s.body}

	if with != nil {
		with(s)
	}

	return s
}

func (s *Scaffold) Body() *Shared[core.Component] {
	return s.body
}

func (s *Scaffold) NavigationComponent() EmbeddedNavigationComponent {
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
	return ora.Scaffold{
		Ptr:                 s.id,
		Type:                ora.ScaffoldT,
		Body:                renderComponentProp(s.body, s.body),
		NavigationComponent: s.navigationComponent.render(),
	}
}
