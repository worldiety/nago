package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"strconv"
)

type RenderContext interface {
	// Window returns the associated Window instance.
	Window() Window

	// MountCallback returns for non-nil funcs a pointer. This pointer is only unique for the current render state.
	// This means, that subsequent calls which result in the same structural ora tree, will have the same
	// pointers. This allows more efficient model deltas. The largest downside is, that an outdated frontend
	// may invoke the wrong callbacks.
	// All callbacks are removed between render calls.
	MountCallback(func()) ora.Ptr

	// Handle returns a unique pointer based on the contents of the given buffer. Note, that for performance reasons
	// the implementation may assume static slices and short circuit based on the slice pointer. It is only guaranteed
	// that the returned pointer is valid during the window lifetime. The first time, a handle is created, the returned
	// flag is true. Also check, if hnd is 0, e.g. due to nil slices. Important: the returned pointers are only valid
	// for the scope lifetime.
	Handle([]byte) (hnd ora.Ptr, created bool)
}

type State[T any] struct {
	id       string
	ptr      ora.Ptr
	value    T
	valid    bool
	observer []func(newValue T)
}

func (s *State[T]) Observe(f func(newValue T)) {
	s.observer = append(s.observer, f)
}

func (s *State[T]) String() string {
	return fmt.Sprintf("%v", s.value)
}

func (s *State[T]) ID() ora.Ptr {
	return s.ptr
}

func (s *State[T]) parse(v any) error {
	switch any(s.value).(type) {
	case bool:
		b, err := strconv.ParseBool(fmt.Sprintf("%v", v))
		if err != nil {
			return err
		}
		s.value = any(b).(T)
	case float64:
		f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		if err != nil {
			return err
		}
		s.value = any(f).(T)
	case int64:
		i, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			return err
		}
		s.value = any(i).(T)
	case ora.Date:
		obj := v.(map[string]interface{})
		var d ora.Date
		d.Day = int(obj["d"].(float64))
		d.Month = int(obj["m"].(float64))
		d.Year = int(obj["y"].(float64))
		s.value = any(d).(T)
	default:
		s.value = any(v).(T)
	}

	for _, f := range s.observer {
		f(s.value)
	}

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
	if s == nil {
		return 0
	}
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
