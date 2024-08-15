package core

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

// A State is held within the root composition (which is a scope)
// and survives a render call. However, after each render cycle,
// the state generation is checked, and unused states are removed because
// their associated views are obviously detached from the tree.
// This also avoids memory leaks by unused states.
// A State is thread safe, however there are deadlock situations,
// when changing observers and values in a nested way, so keep your code simple.
//
// Important difference: a state change will never trigger a rendering by itself, to keep
// render performance up. In most situations (like a frontend event), a rendering will happen automatically,
// but if the domain generates a new state, you have to observe that and issue an invalidation manually.
type State[T any] struct {
	id                    string
	ptr                   ora.Ptr
	value                 T
	valid                 bool
	observer              []func(newValue T)
	destroyedObserver     []func()
	generation            int64
	lastChangedGeneration int64
	mutex                 sync.Mutex
	observerLock          sync.RWMutex
	destroyed             bool
	wnd                   Window
}

// ID returns the window unique state identifier which is internally mapped into an ora.Ptr.
func (s *State[T]) ID() string {
	return s.id
}

// Window returns the window in which this state has been allocated.
func (s *State[T]) Window() Window {
	return s.wnd
}

func (s *State[T]) Observe(f func(newValue T)) {
	s.observerLock.Lock()
	defer s.observerLock.Unlock()

	s.observer = append(s.observer, f)
}

func (s *State[T]) clearObservers() {
	s.observerLock.Lock()
	defer s.observerLock.Unlock()

	clear(s.observer)           // nil out so that GC can collect them
	s.observer = s.observer[:0] // ensure buffer re-usage between re-renderings
}

func (s *State[T]) String() string {
	return fmt.Sprintf("%v", s.value)
}

func (s *State[T]) ptrId() ora.Ptr {
	return s.ptr
}

func (s *State[T]) setGeneration(g int64) {
	atomic.StoreInt64(&s.generation, g)
}

func (s *State[T]) getGeneration() int64 {
	return atomic.LoadInt64(&s.generation)
}

func (s *State[T]) parse(v any) error {
	switch any(s.value).(type) {
	case bool:
		b, err := strconv.ParseBool(fmt.Sprintf("%v", v))
		if err != nil {
			return err
		}

		s.Set(any(b).(T))
	case float64:
		f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		if err != nil {
			return err
		}

		s.Set(any(f).(T))
	case int64:
		i, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			return err
		}

		s.Set(any(i).(T))
	case ora.Date:
		obj := v.(map[string]interface{})
		var d ora.Date
		d.Day = int(obj["d"].(float64))
		d.Month = int(obj["m"].(float64))
		d.Year = int(obj["y"].(float64))

		s.Set(any(d).(T))

	case string:
		s.Set(any(fmt.Sprintf("%v", v)).(T))

	default:
		if t, ok := any(v).(T); ok {
			s.Set(t)
		} else {
			return fmt.Errorf("invalid type %T", v)
		}
	}

	s.observerLock.RLock()
	defer s.observerLock.RUnlock()

	for _, f := range s.observer {
		f(s.value)
	}

	return nil
}

func (s *State[T]) Get() T {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.value
}

func (s *State[T]) Set(v T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	atomic.StoreInt64(&s.lastChangedGeneration, s.getGeneration()) // TODO set this once in the scope_window for better performance
	s.value = v
	s.valid = true
}

func (s *State[T]) addDestroyObserver(fn func()) {
	s.observerLock.Lock()
	defer s.observerLock.Unlock()

	s.destroyedObserver = append(s.destroyedObserver, fn)
}

func (s *State[T]) destroy() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.destroyed {
		// make idempotent
		return
	}

	s.destroyed = true
	s.observerLock.Lock()
	defer s.observerLock.Unlock()

	for _, fn := range s.destroyedObserver {
		fn()
	}
}

func (s *State[T]) isDestroyed() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.destroyed
}

// dirty returns true, if the value has been changed within the current render generation.
func (s *State[T]) dirty() bool {
	return atomic.LoadInt64(&s.lastChangedGeneration) >= s.getGeneration()
}

func (s *State[T]) Ptr() ora.Ptr {
	if s == nil {
		return 0
	}

	return s.ptr
}

// From executes the given func if the State has been
// just initialized and is still invalid.
func (s *State[T]) From(fn func() T) *State[T] {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.valid {
		s.valid = true
		s.value = fn()
	}
	return s
}

func StateOf[T any](wnd Window, id string) *State[T] {
	w := wnd.(*scopeWindow)

	if id == "" {
		panic("empty id is not allowed, consider using AutoState instead")
	}
	some, ok := w.statesById[id]
	if ok {
		if found, ok := some.(*State[T]); ok {
			found.setGeneration(w.generation)
			return found
		}
		var zero T
		slog.Error("restored view state does not match expected state type", "expected", fmt.Sprintf("%T", zero), "got", fmt.Sprintf("%T", some))
	}

	w.lastStatePtrById++
	state := &State[T]{
		wnd:        wnd,
		id:         id,
		ptr:        w.lastStatePtrById,
		valid:      false,
		generation: w.generation,
	}
	w.states[w.lastStatePtrById] = state
	w.statesById[id] = state

	return state
}

// AutoState uses the structural identity to associate
// the actual state in the composition. The implementation
// may change over time to improve reliability or performance.
// Currently, the implementation calculates an identifier based
// on the program counter of the callee, which is quite expensive
// but more stable than just incrementing a naive invocation counter.
//
// Compare that to jetpack compose, which inserts structural identifiers
// into each composable function, see also the article of Leland
// Richardson "Under the hood of Jetpack Compose".
func AutoState[T any](wnd Window) *State[T] {
	// why [3]? Because after the last 3 frames, we have different render entry points (e.g. by post, by timer, by frontend etc)
	// thus using more than the minimum stable 3 frames, we will get lost states, depending on render-cause.
	// Thus, for now, we truncate, but that may be totally wrong either.
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])

	// be as efficient as possible
	id := sha512.Sum512_224((*[2 * unsafe.Sizeof(uintptr(0))]byte)(unsafe.Pointer(&pcs[0]))[:n])
	const debug = false
	if debug {
		frames := runtime.CallersFrames(pcs[:n])
		fmt.Println("---->")
		for {
			frame, more := frames.Next()
			fmt.Printf("%d Function: %s, Line: %d %d =>%s\n", frame.PC, frame.Function, frame.Line, frame.Entry, hex.EncodeToString(id[:]))
			if !more {
				break
			}
		}
		fmt.Println("<----")
	}

	// be as efficient as possible, I know, this is not unicode
	return StateOf[T](wnd, unsafe.String(&id[0], len(id)))
}

// OnAppear executes the given function only once for the given identity state. It triggers an invalidation
// after the fn is eolDone, but if you need to redraw while running, use [Window.Invalidate].
// The identity is asserted by the given id. If empty, an AutoState derived by structural identity is assumed.
// This is similar to Jetpack Compose LaunchedEffect, see also
// https://developer.android.com/develop/ui/compose/side-effects#launchedeffect.
// The given func is executed from a different go routine, to avoid blocking the render thread.
// Important: setting state values is always thread safe, but do NOT set other variables from here, especially
// never global vars or stack-local (which will likely be effect free).
// Spawning another go-routing without blocking, will cancel the given context, as soon as given fn exists, thus
// you need to block it. The context is also cancelled, if the state is destroyed.
//
// It follows the identical lifecycle rules as State. See also [OnAsyncDisappear].
func OnAppear(wnd Window, id string, fn func(ctx context.Context)) {
	var state *State[bool]
	if id == "" {
		state = AutoState[bool](wnd)
	} else {
		state = StateOf[bool](wnd, id)
	}

	state.From(func() bool {
		// even though it is not documented clearly, TheMerovius tells us, that cancelling a context is idempotent
		ctx, cancel := context.WithCancel(wnd.Context())

		state.addDestroyObserver(func() {
			cancel()
		})

		go func() {
			// not sure what to do here: this may mean, that the ctx escaped to another go-routine in fn
			defer cancel()
			fn(ctx)
			//wnd.Invalidate()
		}()

		return true
	})
}

// OnDisappear is executed, once the identified state goes out of scope. Otherwise, the rules of [OnAsyncAppear]
// are applied.
func OnDisappear(wnd Window, id string, fn func(ctx context.Context)) {
	var state *State[bool]
	if id == "" {
		state = AutoState[bool](wnd)
	} else {
		state = StateOf[bool](wnd, id)
	}
	state.From(func() bool {
		ctx, cancel := context.WithCancel(wnd.Context())

		state.addDestroyObserver(func() {
			go func() {
				defer cancel()
				fn(ctx)
				//wnd.Invalidate()
			}()
		})

		return true
	})
}
