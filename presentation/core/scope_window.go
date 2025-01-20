package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/ora"
	"golang.org/x/text/language"
	"io"
	"log/slog"
	"sync"
	"time"
)

var _ Window = (*scopeWindow)(nil)

const maxAutoPtr = 10_000

type declaredBufferKey struct {
	ptr *byte
	len int
}

type scopeWindow struct {
	parent        *Scope
	rootFactory   std.Option[ComponentFactory]
	lastRendering std.Option[ora.Component]
	destroyed     bool
	callbackPtr   ora.Ptr
	callbacks     map[ora.Ptr]func()
	//lastAutoStatePtr      ora.Ptr
	lastStatePtrById      ora.Ptr
	states                map[ora.Ptr]Property
	statesById            map[string]Property
	filesReceiver         map[ora.Ptr]FilesReceiver
	destroyObservers      map[int]func()
	importFilesReceivers  map[string]ImportFilesOptions
	exportFilesReceivers  map[string]ExportFilesOptions
	hnd                   int
	factory               ora.ComponentFactoryId
	navController         *navigationController
	values                Values
	declaredBuffers       map[declaredBufferKey]ora.Ptr
	lastDeclaredBufferPtr ora.Ptr
	isRendering           bool
	generation            int64
	mutex                 sync.Mutex
}

func newScopeWindow(parent *Scope, factory ora.ComponentFactoryId, values Values) *scopeWindow {
	s := &scopeWindow{parent: parent}
	s.callbacks = map[ora.Ptr]func(){}
	s.factory = factory
	s.states = map[ora.Ptr]Property{}
	s.statesById = map[string]Property{}
	s.lastStatePtrById = maxAutoPtr
	s.declaredBuffers = map[declaredBufferKey]ora.Ptr{}
	s.generation = 0

	if values == nil {
		s.values = Values{}
	} else {
		s.values = values
	}

	s.navController = newNavigationController(parent)
	for _, observer := range s.parent.app.onWindowCreatedObservers {
		observer(s)
	}

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
	s.callbackPtr = 0 // make them stable
	//s.lastAutoStatePtr = 0 // make them stable
	//clear(s.states)
	clear(s.filesReceiver)
	//clear(s.destroyObservers) ???
	clear(s.callbacks)

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

func (s *scopeWindow) render() ora.Component {
	s.isRendering = true
	s.generation++
	defer func() {
		s.isRendering = false
		s.removeDetachedStates(s.generation)
	}()

	if !s.rootFactory.Valid {
		panic("invalid root factory")
	}
	s.reset()

	fac := s.rootFactory.Unwrap()
	component := fac(s)
	if component == nil {
		panic(fmt.Errorf("factory '%s' returned a nil component which is not allowed", s.factory))
	}

	tree := component.Render(s)

	// update global scope transient states with the latest render generation.
	// this is used by the ticker to check, if a re-render is required
	for _, property := range s.parent.statesById {
		property.setGeneration(s.generation)
	}

	return tree
}

func (s *scopeWindow) Window() Window {
	return s
}

func (s *scopeWindow) SetColorScheme(scheme ColorScheme) {
	s.parent.Publish(ora.ThemeRequested{
		Type:  ora.ThemeRequestedT,
		Theme: string(scheme),
	})
}

func (s *scopeWindow) Path() NavigationPath {
	return NavigationPath(s.factory)
}

func (s *scopeWindow) AddDestroyObserver(fn func()) (removeObserver func()) {
	if s.destroyObservers == nil {
		s.destroyObservers = make(map[int]func())
	}
	s.hnd++
	myHnd := s.hnd
	s.destroyObservers[myHnd] = fn
	return func() {
		delete(s.destroyObservers, myHnd)
	}
}

func (s *scopeWindow) Invalidate() {
	s.Execute(func() {
		if s.destroyed {
			return
		}
		s.parent.forceRender(0)
	})

}

func (s *scopeWindow) destroy() {
	s.destroyed = true

	for _, property := range s.states {
		property.clearObservers()
		property.destroy()
	}

	for _, f := range s.destroyObservers {
		f()
	}
	clear(s.destroyObservers)

}

func (s *scopeWindow) Handle(buf []byte) (ora.Ptr, bool) {
	if len(buf) == 0 {
		return 0, false
	}

	key := declaredBufferKey{
		ptr: &buf[0],
		len: len(buf),
	}

	ptr, ok := s.declaredBuffers[key]
	if ok {
		return ptr, false
	}

	s.lastDeclaredBufferPtr++
	s.declaredBuffers[key] = s.lastDeclaredBufferPtr
	return s.lastDeclaredBufferPtr, true
}

func (s *scopeWindow) MountCallback(f func()) ora.Ptr {
	if f == nil {
		return 0
	}
	s.callbackPtr++
	s.callbacks[s.callbackPtr] = f

	return s.callbackPtr
}

func (s *scopeWindow) Application() *Application {
	return s.parent.app
}

func (s *scopeWindow) UpdateSubject(subject auth.Subject) {
	if subject == nil {
		subject = auth.InvalidSubject{}
	}

	s.parent.subject.SetValue(subject)
}

func (s *scopeWindow) AsURI(open func() (io.Reader, error)) (URI, error) {
	if s.destroyed {
		return "", nil
	}

	if callback := s.parent.app.onShareStream; callback != nil {
		return callback(s.parent, open)
	}

	return "", fmt.Errorf("no share stream platform adapter has been configured")
}

func (s *scopeWindow) ImportFiles(options ImportFilesOptions) {
	if s.destroyed {
		return
	}

	if s.isRendering {
		panic("you must not call ImportFiles from the render loop, only from action or post is allowed")
	}

	if options.OnCompletion == nil {
		panic("OnCompletion is required")
	}

	if options.ID == "" {
		s.callbackPtr++
		options.ID = fmt.Sprintf("auto-%d", s.callbackPtr)
	}

	if s.importFilesReceivers == nil {
		s.importFilesReceivers = map[string]ImportFilesOptions{}
	}

	if options.MaxBytes == 0 {
		options.MaxBytes = 1024 * 1024 * 512 // defaults to 512MiB
	}

	s.importFilesReceivers[options.ID] = options

	s.parent.Publish(ora.FileImportRequested{
		Type:             ora.FileImportRequestedT,
		ID:               options.ID,
		ScopeID:          string(s.parent.id),
		Multiple:         options.Multiple,
		MaxBytes:         options.MaxBytes,
		AllowedMimeTypes: options.AllowedMimeTypes,
	})
}

func (s *scopeWindow) ExportFiles(options ExportFilesOptions) {
	if s.destroyed {
		return
	}

	if s.isRendering {
		panic("you must not call SendFiles from the render loop, only from action or post is allowed")
	}

	if options.ID == "" {
		s.callbackPtr++
		options.ID = fmt.Sprintf("auto-%d", s.callbackPtr)
	}

	if s.exportFilesReceivers == nil {
		s.exportFilesReceivers = map[string]ExportFilesOptions{}
	}

	s.exportFilesReceivers[options.ID] = options

	if callback := s.parent.app.onSendFiles; callback != nil {
		if err := callback(s.parent, options); err != nil {
			slog.Error("cannot export files", "err", err)
		}

		return
	}

	slog.Error("no send files platform adapter has been configured")
}

func (s *scopeWindow) Execute(task func()) {
	if s.destroyed {
		return
	}

	s.parent.eventLoop.Post(task)
	s.parent.eventLoop.Tick()
}

func (s *scopeWindow) Info() WindowInfo {
	return s.parent.windowInfo
}

func (s *scopeWindow) Navigation() Navigation {
	return s.navController
}

func (s *scopeWindow) Values() Values {
	return s.values
}

func (s *scopeWindow) Subject() auth.Subject {
	return s.parent.subject.Value()
}

func (s *scopeWindow) Context() context.Context {
	return s.parent.ctx
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
