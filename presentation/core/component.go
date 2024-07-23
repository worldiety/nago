package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
)

type RenderContext interface {
	Window() Window
	// MountCallback returns for non-nil funcs a pointer. This pointer is only unique for the current render state.
	// This means, that subsequent calls which result in the same structural ora tree, will have the same
	// pointers. This allows more efficient model deltas. The largest downside is, that an outdated frontend
	// may invoke the wrong callbacks.
	// All callbacks are removed between render calls.
	MountCallback(func()) ora.Ptr
}

type State[T any] struct {
	id    string
	ptr   ora.Ptr
	value T
	valid bool
}

func (s *State[T]) ID() ora.Ptr {
	return s.ptr
}

func (s *State[T]) Parse(v string) error {
	s.value = any(v).(T)
	return nil
}

func (s *State[T]) Get() T {
	return s.value
}

func (s *State[T]) Set(v T) {
	s.value = v
	s.valid = true
}

func (s *State[T]) Ptr() ora.Ptr {
	return s.ptr
}

func StateWithID[T any](wnd Window, id string) *State[T] {
	w := wnd.(*scopeWindow)

	if id == "" {
		panic("empty id is not allowed, consider using AutoState instead")
	}
	some, ok := w.statesById[id]
	if ok {
		if found, ok := some.(*State[T]); ok {
			return found
		}
		var zero T
		slog.Error("restored view state does not match expected state type", "expected", fmt.Sprintf("%T", zero), "got", fmt.Sprintf("%T", some))
	}

	w.lastStatePtrById++
	state := &State[T]{
		id:    id,
		ptr:   w.lastStatePtrById,
		valid: false,
	}
	w.states[w.lastStatePtrById] = state
	w.statesById[id] = state

	return state
}

func AutoState[T any](wnd Window) *State[T] {
	w := wnd.(*scopeWindow)
	w.lastAutoStatePtr++

	if w.lastAutoStatePtr >= maxAutoPtr {
		panic("auto state overflow, you have to many auto states. Having so many states indicates to large UIs. Reduce or use StateWithID instead")
	}

	fmt.Println("autostate ptr created", w.lastAutoStatePtr)
	some, ok := w.states[w.lastAutoStatePtr]
	if ok {
		if found, ok := some.(*State[T]); ok {
			return found
		}
		var zero T
		slog.Error("restored view state does not match expected state type", "expected", fmt.Sprintf("%T", zero), "got", fmt.Sprintf("%T", some))
	}

	state := &State[T]{
		id:    "",
		ptr:   w.lastAutoStatePtr,
		valid: false,
	}
	w.states[w.lastAutoStatePtr] = state

	return state

}

type View interface {
	//Padding(padding ora.Padding) View
	Render(RenderContext) ora.Component
}

type DecoredView interface {
	View
	Padding(padding ora.Padding) DecoredView
	Frame(frame ora.Frame) DecoredView
	Border(border ora.Border) DecoredView
	Visible(visible bool) DecoredView
	// AccessibilityLabel is used to help screen readers, see guidelines, when use them.
	// https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel(label string) DecoredView
}

type ViewPadding struct {
	parent  View
	padding *ora.Padding
}

func NewViewPadding(parent View, padding *ora.Padding) ViewPadding {
	return ViewPadding{parent: parent, padding: padding}
}

func (p ViewPadding) Top(pad ora.Length) View {
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) All(pad ora.Length) View {
	p.padding.Left = pad
	p.padding.Right = pad
	p.padding.Bottom = pad
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) Vertical(pad ora.Length) View {
	p.padding.Bottom = pad
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) Horizontal(pad ora.Length) View {
	p.padding.Left = pad
	p.padding.Right = pad
	return p.parent
}
