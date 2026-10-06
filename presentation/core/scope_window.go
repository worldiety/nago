// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/logging"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
)

var _ Window = (*scopeWindow)(nil)

var globalListenerPtr atomic.Int64

type declaredBufferKey struct {
	ptr *byte
	len int
}

type scopeWindow struct {
	parent        *Scope
	rootFactory   std.Option[ComponentFactory]
	lastRendering std.Option[proto.Component]
	// destroyed is written by the event loop but read from any goroutine, e.g. by Invalidate.
	destroyed atomic.Bool
	// callbacks holds the callbacks of the tree rendered last, see MountCallback.
	callbacks callbackSegment
	// former holds the keys of recently superseded trees, oldest first, see [MountKeyedCallback].
	former []formerSegment
	// spareKeys is a released key slice of a former tree, reused by the next tree.
	spareKeys        []CallbackKey
	states           map[proto.Ptr]Property
	statesById       map[string]Property
	resetObservers   map[int]func()
	destroyObservers map[int]func()
	hnd              int
	factory          proto.RootViewID
	navController    *navigationController
	values           atomic.Pointer[Values]
	isRendering      bool
	generation       int64
	// dirtyGeneration is set to the current generation whenever any [State] which belongs to this window
	// is mutated (see markStateDirty). It allows the frame ticker to decide in O(1) whether a re-render is
	// required, instead of iterating over all states. It is accessed atomically because the frame ticker reads
	// it concurrently to the render loop.
	dirtyGeneration int64
	// lastDirtyState is the ID of the state which has been marked dirty last, see checkRenderLoop.
	lastDirtyState atomic.Pointer[string]
	// the render loop detection, only for the event loop
	renderLoopCount  int
	renderLoopSince  time.Time
	renderLoopWarned bool
	mutex            sync.Mutex
	clipboard        *clipboardController
	asyncCallbacks   concurrent.RWMap[proto.Ptr, asyncCallback]
}

func (s *scopeWindow) Clipboard() Clipboard {
	return s.clipboard
}

func newScopeWindow(parent *Scope, factory proto.RootViewID, values Values) *scopeWindow {
	s := &scopeWindow{parent: parent}
	s.callbacks.restart(parent.ids.callback)
	s.factory = factory
	s.states = map[proto.Ptr]Property{}
	s.statesById = map[string]Property{}
	// continue with the generation of the scope, otherwise the transient states of the scope, which carry the
	// generation of the former window, would be considered dirty until this window has caught up.
	s.generation = parent.generation.Load()
	s.dirtyGeneration = -1

	if values == nil {
		values = Values{}
	}

	s.values.Store(&values)

	s.navController = newNavigationController(parent)
	for _, observer := range s.parent.app.onWindowCreatedObservers {
		observer(s)
	}

	s.clipboard = newClipboardController(s)
	return s
}

func (s *scopeWindow) Session() session.UserSession {
	return *s.parent.virtualSession.Load()
}

func (s *scopeWindow) setFactory(view ComponentFactory) {
	if view == nil {
		s.rootFactory = std.None[ComponentFactory]()
		return
	}

	s.rootFactory = std.Some(view)
}

func (s *scopeWindow) reset() {
	// Callbacks are only valid for the tree which has been rendered last. Their pointers are unique within the
	// scope and never reused, so that a stale tree of the frontend can never invoke a callback of a newer tree,
	// e.g. the second click of a double click which would otherwise hit whatever is at the same position now.
	// Only a call with a key may be redirected, see [MountKeyedCallback].
	s.supersedeCallbacks(time.Now())

	// note that we are not clearing the async callbacks here to survive any render cycle.
	// see also the core.DestroyObserverOption to distinguish between reset and destroy life cycle states.
	for _, f := range s.resetObservers {
		f()
	}
	clear(s.resetObservers)

	for _, property := range s.states {
		property.clearObservers()
	}
}

func (s *scopeWindow) removeDetachedStates(currentGeneration int64) {
	for id, property := range s.statesById {
		if property.getGeneration() < currentGeneration {
			delete(s.statesById, id)
			delete(s.states, property.ptrId())
			property.destroy()

			//slog.Info("purged unused state", "id", id, "expected", currentGeneration, "has", property.getGeneration())
		}
	}
}

// generationOf returns the current render generation. It is read atomically, because the frame ticker
// (see hasDirtyStates) may read it concurrently to the render loop.
func (s *scopeWindow) generationOf() int64 {
	return atomic.LoadInt64(&s.generation)
}

// markStateDirty records that a [State] belonging to this window has been mutated within the current
// generation. This is the O(1) replacement for iterating over all states in the frame ticker.
// After the next render increments the generation, dirtyGeneration is implicitly older than the generation
// again, thus the window is considered clean until the next mutation.
func (s *scopeWindow) markStateDirty(id *string) {
	s.lastDirtyState.Store(id)

	// the marker must never move backwards: another goroutine may have loaded an older generation before a
	// render started and would otherwise overwrite a mutation which happened during that render.
	generation := atomic.LoadInt64(&s.generation)
	for {
		marked := atomic.LoadInt64(&s.dirtyGeneration)
		if marked >= generation || atomic.CompareAndSwapInt64(&s.dirtyGeneration, marked, generation) {
			return
		}
	}
}

// A render loop is reported, if a state has been changed during each render for at least renderLoopCount
// renders and renderLoopDuration.
const (
	renderLoopCount    = 50
	renderLoopDuration = 5 * time.Second
)

// checkRenderLoop is called after each render with the information, whether any state is dirty again. If
// that is always the case, a view changes a state during its render, which causes the next render and so
// on. Such a window renders at the rate of the frame ticker and discards the actions of the user, because they
// refer to callbacks of outdated trees. A goroutine which changes a state while a render is running may
// cause the same observation by chance, but not for such a long streak.
// Only for the event loop.
func (s *scopeWindow) checkRenderLoop(dirty bool) {
	if !dirty {
		s.renderLoopCount = 0
		return
	}

	if s.renderLoopCount == 0 {
		s.renderLoopSince = time.Now()
	}

	s.renderLoopCount++
	if s.renderLoopWarned || s.renderLoopCount < renderLoopCount || time.Since(s.renderLoopSince) < renderLoopDuration {
		return
	}

	s.renderLoopWarned = true
	var state string
	if id := s.lastDirtyState.Load(); id != nil {
		state = *id
	}

	slog.Warn("window renders in an endless loop, because a state changes during each render, e.g. by setting a changing value or by Invalidate within a render", "scope", logging.Secret(string(s.parent.id)), "view", s.factory, "renders", s.renderLoopCount, "since", s.renderLoopSince, "lastChangedState", state)
}

// hasDirtyStates reports in O(1) whether any [State] of this window has been mutated since the last render.
// It intentionally does not cover the scope-level transient states, which are still checked separately.
func (s *scopeWindow) hasDirtyStates() bool {
	return atomic.LoadInt64(&s.dirtyGeneration) >= atomic.LoadInt64(&s.generation)
}

func (s *scopeWindow) render() proto.Component {
	s.isRendering = true
	generation := s.parent.generation.Add(1)
	atomic.StoreInt64(&s.generation, generation)
	defer func() {
		s.isRendering = false
		s.removeDetachedStates(generation)
	}()

	if s.rootFactory.IsNone() {
		panic("invalid root factory")
	}
	s.reset()

	fac := s.rootFactory.Unwrap()
	component := fac(s)
	if component == nil {
		panic(fmt.Errorf("factory '%s' returned a nil component which is not allowed", s.factory))
	}

	tree := component.Render(s)

	// Mark the transient states as rendered only now, thus a change during the render does not cause another
	// render. This is intentional, because a transient state has no equality check, and it is set e.g. by a
	// banner which shows a message within its render. Note that for the same reason, a change from another
	// goroutine during the render is not rendered until the next render, see [State.Set].
	s.parent.setTransientGeneration(generation)

	return tree
}

func (s *scopeWindow) Window() Window {
	return s
}

func (s *scopeWindow) SetColorScheme(scheme ColorScheme) {
	s.parent.Publish(&proto.ThemeRequested{
		Theme: proto.ThemeID(scheme.String()),
	})
}

func (s *scopeWindow) SetFonts(fonts Fonts) {
	faces := proto.FontFaces{}
	for _, face := range fonts.Faces {
		faces = append(faces, proto.FontFace{
			Family: proto.Str(face.Family),
			Style:  proto.Str(face.Style),
			Weight: proto.Str(face.Weight),
			Source: proto.URI(face.Source),
		})
	}

	s.parent.Publish(&proto.FontsRequested{
		Fonts: proto.Fonts{
			DefaultFontFace: proto.Str(fonts.DefaultFont),
			Faces:           faces,
		},
	})
}

func (s *scopeWindow) Path() NavigationPath {
	return NavigationPath(s.factory)
}

func (s *scopeWindow) AddDestroyObserver(fn func(), opts ...DestroyObserverOption) (removeObserver func()) {
	target := DestroyOnReset
	if len(opts) > 0 {
		target = opts[0]
	}

	var observers map[int]func()
	switch target {
	case DestroyOnReset:
		if s.resetObservers == nil {
			s.resetObservers = make(map[int]func())
		}

		observers = s.resetObservers
	case DestroyOnClose:
		if s.destroyObservers == nil {
			s.destroyObservers = make(map[int]func())
		}

		observers = s.destroyObservers
	}

	s.hnd++
	myHnd := s.hnd
	observers[myHnd] = fn
	return func() {
		delete(observers, myHnd)
	}
}

func (s *scopeWindow) Invalidate() {
	s.Execute(func() {
		if s.destroyed.Load() {
			return
		}
		s.parent.forceRender(0)
	})

}

func (s *scopeWindow) destroy() {
	s.destroyed.Store(true)
	s.forgetFormer()

	for _, property := range s.states {
		property.clearObservers()
		property.destroy()
	}

	for _, f := range s.resetObservers {
		f()
	}
	clear(s.resetObservers)

	for _, f := range s.destroyObservers {
		f()
	}
	clear(s.destroyObservers)

	s.asyncCallbacks.Clear()
}

func (s *scopeWindow) MountCallback(f func()) proto.Ptr {
	if f == nil {
		return 0
	}

	ptr := s.callbacks.mount(f)
	s.parent.ids.callback = s.callbacks.next()

	return ptr
}

// dropCallbacks releases the callbacks of the current tree. The next tree continues with the pointers after
// them.
func (s *scopeWindow) dropCallbacks() {
	next := s.callbacks.next()
	if next >= maxCallbackPtr {
		// practically unreachable: at a million callbacks per second, this takes centuries
		slog.Error("the callback pointers of the scope are exhausted, destroying the scope", "scope", logging.Secret(string(s.parent.id)))
		s.parent.Destroy()
	}

	s.callbacks.restart(next)
	s.parent.ids.callback = next
}

func (s *scopeWindow) Application() *Application {
	return s.parent.app
}

func (s *scopeWindow) UpdateSubject(subject auth.Subject) {
	if subject == nil {
		subject = s.parent.app.getAnonUser()
	}

	// nothing rendered for the former subject may be redirected any more
	s.discardKeys()

	if setter, ok := subject.(subjectLanguageSetter); ok {
		setter.SetBundle(s.Bundle())
		setter.SetLanguage(s.Locale())
	}

	s.parent.subject.SetValue(subject)
}

func (s *scopeWindow) Logout() error {
	if _, err := s.parent.app.logoutSession(s.Session().ID()); err != nil {
		return err
	}

	s.UpdateSubject(nil)

	return nil
}

func (s *scopeWindow) AsURI(open func() (io.Reader, error)) (URI, error) {
	if s.destroyed.Load() {
		return "", nil
	}

	if callback := s.parent.app.onShareStream; callback != nil {
		return callback(s.parent, open)
	}

	return "", fmt.Errorf("no share stream platform adapter has been configured")
}

func (s *scopeWindow) ImportFiles(options ImportFilesOptions) {
	if s.destroyed.Load() {
		return
	}

	if s.isRendering {
		panic("you must not call ImportFiles from the render loop, only from action or post is allowed")
	}

	if options.OnCompletion == nil {
		panic("OnCompletion is required")
	}

	if options.ID == "" {
		options.ID = fmt.Sprintf("auto-%d", s.parent.ids.nextFile())
	}

	if options.MaxBytes == 0 {
		options.MaxBytes = 1024 * 1024 * 512 // defaults to 512MiB
	}

	s.parent.putImportFiles(options)

	s.parent.Publish(&proto.FileImportRequested{
		ID:               proto.Str(options.ID),
		ScopeID:          proto.Str(s.parent.id),
		Multiple:         proto.Bool(options.Multiple),
		MaxBytes:         proto.Uint(options.MaxBytes),
		AllowedMimeTypes: intoStrSlice[string, proto.Str](options.AllowedMimeTypes),
	})
}

func (s *scopeWindow) ExportFiles(options ExportFilesOptions) {
	if s.destroyed.Load() {
		return
	}

	if s.isRendering {
		panic("you must not call SendFiles from the render loop, only from action or post is allowed")
	}

	if options.ID == "" {
		options.ID = fmt.Sprintf("auto-%d", s.parent.ids.nextFile())
	}

	s.parent.putExportFiles(options)

	if callback := s.parent.app.onSendFiles; callback != nil {
		if err := callback(s.parent, options); err != nil {
			slog.Error("cannot export files", "err", err)
		}

		return
	}

	slog.Error("no send files platform adapter has been configured")
}

func (s *scopeWindow) Execute(task func()) {
	if s.destroyed.Load() {
		return
	}

	s.parent.eventLoop.Post(task)
}

func (s *scopeWindow) Info() WindowInfo {
	return s.parent.windowInfo
}

func (s *scopeWindow) Navigation() Navigation {
	return s.navController
}

func (s *scopeWindow) Values() Values {
	ptrV := s.values.Load()
	if ptrV == nil {
		return Values{}
	}

	return *ptrV
}

func (s *scopeWindow) Subject() auth.Subject {
	return s.parent.subject.Value()
}

func (s *scopeWindow) Context() context.Context {
	return s.Subject().Context()
}

func (s *scopeWindow) Authenticate() {
	// TODO ????
}

func (s *scopeWindow) Locale() language.Tag {
	return s.parent.locale
}

func (s *scopeWindow) Location() *time.Location {
	return s.parent.location
}

func (s *scopeWindow) Bundle() *i18n.Bundle {
	return s.parent.bundle
}

func (s *scopeWindow) Post(fn func()) bool {
	s.mutex.Lock()
	p := s.parent
	s.mutex.Unlock()

	if p != nil {
		return p.eventLoop.Post(fn)
	}

	return false
}

func (s *scopeWindow) PostDelayed(fn func(), delay time.Duration) {
	time.AfterFunc(delay, func() {
		s.Post(fn)
	})
}

func (s *scopeWindow) RequestFocus(id string) {
	AsyncCall(s, &proto.CallRequestFocus{ID: proto.Str(id)}, nil)
}

func (s *scopeWindow) AddInputListener(elemID string, fn func(evt InputEvent), opts ...InputListenerOption) (close func()) {
	if len(opts) == 0 {
		opts = []InputListenerOption{
			InputEventPointerDown,
			InputEventPointerUp,
			InputEventPointerMove,
			InputEventPointerCancel,
			InputEventKeyDown,
			InputEventKeyUp,
		}
	}
	hnd := globalListenerPtr.Add(1)
	protoTypes := make(proto.InputEventTypes, 0, len(opts))
	destroyOpt := DestroyOnReset
	for _, t := range opts {
		if t, ok := t.(InputEventType); ok {
			protoTypes = append(protoTypes, proto.InputEventType(t))
		}

		if t, ok := t.(DestroyObserverOption); ok {
			destroyOpt = t
		}
	}

	cancel := asyncCall(
		s,
		&proto.RegisterInputEventListener{
			Id:     proto.Str(elemID),
			Handle: proto.Uint(hnd),
			Types:  protoTypes,
		},
		func(ret proto.CallRet) {
			if evt, ok := ret.(*proto.InputEvent); ok {
				fn(InputEvent{
					Type: InputEventType(evt.Type),
					X:    float64(evt.X),
					Y:    float64(evt.Y),
					Code: string(evt.Code),
				})
			}
		},
		true,
	)

	closer := func() {
		AsyncCall(s, &proto.UnregisterInputEventListener{Handle: proto.Uint(hnd)}, nil)
		cancel()
	}

	s.AddDestroyObserver(closer, destroyOpt)
	return closer
}
