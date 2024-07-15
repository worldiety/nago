package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type FlexContainer struct {
	id               ora.Ptr
	elements         *SharedList[core.View]
	elementSize      EmbeddedElementSize
	orientation      EmbeddedOrientation
	contentAlignment EmbeddedContentAlignment
	itemsAlignment   EmbeddedItemsAlignment
	properties       []core.Property
	maxWidth         *Shared[ora.ElementSize]
	visible          Bool
}

func NewFlexContainer(with func(flex *FlexContainer)) *FlexContainer {
	f := &FlexContainer{
		id:               nextPtr(),
		elements:         NewSharedList[core.View]("elements"),
		elementSize:      NewShared[ElementSize]("elementSize"),
		orientation:      NewShared[Orientation]("orientation"),
		contentAlignment: NewShared[ContentAlignment]("contentAlignment"),
		itemsAlignment:   NewShared[ItemsAlignment]("itemsAlignment"),
		visible:          NewShared[bool]("visible"),
		maxWidth:         NewShared[ora.ElementSize]("maxWidth"),
	}

	f.properties = []core.Property{f.elements, f.elementSize, f.orientation, f.contentAlignment, f.itemsAlignment, f.maxWidth}

	// the container is otherwise in an undefined state, so lets define it
	f.orientation.Set(ora.OrientationHorizontal)
	f.contentAlignment.Set(ora.ContentCenter)
	f.itemsAlignment.Set(ora.ItemsStretch)
	f.elementSize.Set(ora.ElementSizeAuto)
	f.visible.Set(true)
	f.maxWidth.Set(ora.ElementSizeAuto)

	if with != nil {
		with(f)
	}

	return f
}

func (f *FlexContainer) Append(children ...core.View) *FlexContainer {
	f.Elements().Append(children...)
	return f
}

func (f *FlexContainer) Children() *SharedList[core.View] {
	return f.elements
}

func (f *FlexContainer) MaxWidth() *Shared[ora.ElementSize] {
	return f.maxWidth
}

// deprecated: this is called Children in other ViewGroups
func (f *FlexContainer) Elements() *SharedList[core.View] {
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

func (f *FlexContainer) ItemsAlignment() EmbeddedItemsAlignment {
	return f.itemsAlignment
}

func (f *FlexContainer) ID() ora.Ptr {
	return f.id
}

func (f *FlexContainer) Type() string {
	return "FlexContainer"
}

func (f *FlexContainer) Visible() Bool {
	return f.visible
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
		ItemsAlignment:   f.itemsAlignment.render(),
		Visible:          f.visible.render(),
		MaxWidth:         f.maxWidth.render(),
	}
}
