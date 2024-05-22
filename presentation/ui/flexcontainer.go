package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type FlexContainer struct {
	id               ora.Ptr
	elements         *SharedList[core.Component]
	elementSize      EmbeddedElementSize
	orientation      EmbeddedOrientation
	contentAlignment EmbeddedContentAlignment
	properties       []core.Property
}

func NewFlexContainer(with func(flexContainer *FlexContainer)) *FlexContainer {
	f := &FlexContainer{
		id:               nextPtr(),
		elements:         NewSharedList[core.Component]("elements"),
		elementSize:      NewShared[ElementSize]("elementSize"),
		orientation:      NewShared[Orientation]("orientation"),
		contentAlignment: NewShared[ContentAlignment]("contentAlignment"),
	}

	f.properties = []core.Property{f.elements, f.elementSize, f.orientation, f.contentAlignment}

	if with != nil {
		with(f)
	}

	return f
}

func (f *FlexContainer) Elements() *SharedList[core.Component] {
	return f.elements
}

func (f *FlexContainer) ElementSize() EmbeddedElementSize {
	return f.elementSize
}

func (f *FlexContainer) Orientation() EmbeddedOrientation {
	return f.orientation
}

func (f *FlexContainer) ContentAlignment() EmbeddedContentAlignment {
	return f.contentAlignment
}

func (f *FlexContainer) ID() ora.Ptr {
	return f.id
}

func (f *FlexContainer) Type() string {
	return "FlexContainer"
}

func (f *FlexContainer) Properties(yield func(core.Property) bool) {
	for _, property := range f.properties {
		if !yield(property) {
			return
		}
	}
}

func (f *FlexContainer) Render() ora.Component {
	return ora.FlexContainer{
		Ptr:              f.id,
		Type:             ora.FlexContainerT,
		Elements:         renderSharedListComponents(f.elements),
		ElementSize:      f.elementSize.render(),
		Orientation:      f.orientation.render(),
		ContentAlignment: f.contentAlignment.render(),
	}
}
