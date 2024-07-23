package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/ora"
	"golang.org/x/text/language"
	"io"
	"time"
)

var _ Window = (*scopeWindow)(nil)

const maxAutoPtr = 10_000

type scopeWindow struct {
	parent           *Scope
	rootFactory      std.Option[ComponentFactory]
	lastRendering    std.Option[ora.Component]
	destroyed        bool
	callbackPtr      ora.Ptr
	callbacks        map[ora.Ptr]func()
	lastAutoStatePtr ora.Ptr
	lastStatePtrById ora.Ptr
	states           map[ora.Ptr]Property
	statesById       map[string]Property
	filesReceiver    map[ora.Ptr]FilesReceiver
	destroyObservers map[int]func()
	hnd              int
	factory          ora.ComponentFactoryId
	navController    *NavigationController
	values           Values
}

func newScopeWindow(parent *Scope, factory ora.ComponentFactoryId, values Values) *scopeWindow {
	s := &scopeWindow{parent: parent}
	s.callbacks = map[ora.Ptr]func(){}
	s.factory = factory
	s.states = map[ora.Ptr]Property{}
	s.statesById = map[string]Property{}
	s.lastStatePtrById = maxAutoPtr
	if values == nil {
		s.values = Values{}
	}

	s.navController = NewNavigationController(parent)
	for _, observer := range s.parent.app.onWindowCreatedObservers {
		observer(s)
	}

	return s
}

func (s *scopeWindow) setFactory(view ComponentFactory) {
	if view == nil {
		s.rootFactory = std.None[ComponentFactory]()
		return
	}

	s.rootFactory = std.Some(view)
}

func (s *scopeWindow) reset() {
	s.callbackPtr = 0      // make them stable
	s.lastAutoStatePtr = 0 // make them stable
	//clear(s.states)
	clear(s.filesReceiver)
	//clear(s.destroyObservers) ???
	clear(s.callbacks)
}

func (s *scopeWindow) render() ora.Component {

	if !s.rootFactory.Valid {
		panic("invalid root factory")
	}
	s.reset()

	fac := s.rootFactory.Unwrap()
	component := fac(s)
	if component == nil {
		panic(fmt.Errorf("factory '%s' returned a nil component which is not allowed", s.factory))
	}

	return component.Render(s)
}

func (s *scopeWindow) Window() Window {
	return s
}

func (s *scopeWindow) Factory() ora.ComponentFactoryId {
	return s.factory
}

func (s *scopeWindow) AddDestroyObserver(fn func()) (removeObserver func()) {
	s.hnd++
	myHnd := s.hnd
	s.destroyObservers[myHnd] = fn
	return func() {
		delete(s.destroyObservers, myHnd)
	}
}

func (s *scopeWindow) Invalidate() {
	if s.destroyed {
		return
	}
	s.parent.forceRender()
}

func (s *scopeWindow) destroy() {
	s.destroyed = true
	for _, f := range s.destroyObservers {
		f()
	}
	clear(s.destroyObservers)
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
	s.parent.subject.SetValue(subject)
}

func (s *scopeWindow) AsURI(open func() (io.Reader, error)) (ora.URI, error) {
	if s.destroyed {
		return "", nil
	}

	if callback := s.parent.app.onShareStream; callback != nil {
		return callback(s.parent, open)
	}

	return "", fmt.Errorf("no share stream platform adapter has been configured")
}

func (s *scopeWindow) SendFiles(it iter.Seq2[File, error]) error {
	if s.destroyed {
		return nil
	}

	if callback := s.parent.app.onSendFiles; callback != nil {
		return callback(s.parent, it)
	}

	return fmt.Errorf("no send files platform adapter has been configured")
}

func (s *scopeWindow) Execute(task func()) {
	if s.destroyed {
		return
	}

	s.parent.eventLoop.Post(task)
	s.parent.eventLoop.Tick()
}

func (s *scopeWindow) Info() ora.WindowInfo {
	return s.parent.windowInfo
}

func (s *scopeWindow) Navigation() *NavigationController {
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

func (s *scopeWindow) SessionID() SessionID {
	return s.parent.sessionID
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
