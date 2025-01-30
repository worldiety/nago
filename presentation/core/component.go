package core

import (
	"go.wdy.de/nago/presentation/proto"
)

const Debug = false // TODO make be a compile time flagged const

type RenderContext interface {
	// Window returns the associated Window instance.
	Window() Window

	// MountCallback returns for non-nil funcs a pointer. This pointer is only unique for the current render state.
	// This means, that subsequent calls which result in the same structural ora tree, will have the same
	// pointers. This allows more efficient model deltas. The largest downside is, that an outdated frontend
	// may invoke the wrong callbacks.
	// All callbacks are removed between render calls.
	MountCallback(func()) proto.Ptr

	// Handle returns a unique pointer based on the contents of the given buffer. Note, that for performance reasons
	// the implementation may assume static slices and short circuit based on the slice pointer. It is only guaranteed
	// that the returned pointer is valid during the window lifetime. The first time, a handle is created, the returned
	// flag is true. Also check, if hnd is 0, e.g. due to nil slices. Important: the returned pointers are only valid
	// for the scope lifetime.
	Handle([]byte) (hnd proto.Ptr, created bool)
}

type RenderNode = proto.Component

type View interface {
	Render(RenderContext) RenderNode
}

type ViewPadding struct {
	parent  View
	padding *proto.Padding
}

func NewViewPadding(parent View, padding *proto.Padding) ViewPadding {
	return ViewPadding{parent: parent, padding: padding}
}

func (p ViewPadding) Top(pad proto.Length) View {
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) All(pad proto.Length) View {
	p.padding.Left = pad
	p.padding.Right = pad
	p.padding.Bottom = pad
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) Vertical(pad proto.Length) View {
	p.padding.Bottom = pad
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) Horizontal(pad proto.Length) View {
	p.padding.Left = pad
	p.padding.Right = pad
	return p.parent
}
