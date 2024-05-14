package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type NavigationComponent struct {
	id         ora.Ptr
	logo       EmbeddedSVG
	menu       *SharedList[*MenuEntry]
	alignment  EmbeddedAlignment
	properties []core.Property
}

func NewNavigationComponent(with func(navigationComponent *NavigationComponent)) *NavigationComponent {
	n := &NavigationComponent{
		id:        nextPtr(),
		logo:      NewShared[SVGSrc]("logo"),
		menu:      NewSharedList[*MenuEntry]("menu"),
		alignment: NewShared[Alignment]("alignment"),
	}

	n.properties = []core.Property{n.logo, n.menu, n.alignment}

	if with != nil {
		with(n)
	}

	return n
}

func (n *NavigationComponent) Logo() EmbeddedSVG {
	return n.logo
}

func (n *NavigationComponent) Menu() *SharedList[*MenuEntry] {
	return n.menu
}

func (n *NavigationComponent) Alignment() EmbeddedAlignment {
	return n.alignment
}

func (n *NavigationComponent) ID() ora.Ptr {
	return n.id
}

func (n *NavigationComponent) Type() string {
	return "NavigationComponent"
}

func (n *NavigationComponent) Properties(yield func(core.Property) bool) {
	for _, property := range n.properties {
		if !yield(property) {
			return
		}
	}
}

func (n *NavigationComponent) Render() ora.Component {
	return n.renderNavigationComponent()
}

func (n *NavigationComponent) renderNavigationComponent() ora.NavigationComponent {
	return ora.NavigationComponent{
		Ptr:       n.id,
		Type:      ora.NavigationComponentT,
		Logo:      n.logo.render(),
		Menu:      renderSharedListMenuEntries(n.menu),
		Alignment: n.alignment.render(),
	}
}

type ScaffoldTopBar struct {
	Left  *Shared[core.Component]
	Mid   *Shared[core.Component]
	Right *Shared[core.Component]
}
