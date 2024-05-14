package core

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Destroyable interface {
	Destroy()
}

type allocatedComponent struct {
	Window      *scopeWindow
	Component   Component
	RenderState *RenderState
}

type ComponentFactory func(Window, ora.NewComponentRequested) Component

// A Scope manage its own area of associated pointers. A Pointer must be only unique per Scope.
// Resolving or keeping pointers outside a scope is inherently unsafe (e.g. a lookup map).
// A Scope can hold an arbitrary amount of components, created by an arbitrary amount of factories.
// This is intentional. E.g. a mobile app can create multiple instances of the same (or different)
// "pages" and those pages may communicate immediately with each other (observer etc.) without
// causing race conditions.
// Each Scope has a single event loop to guarantee race free event processing.
// A Scope is probably a single window.
// To allow interacting between different scopes, we may either go the full serializing message route or
// as a cheap alternative, replace all event loops with the same looper instance.
// However, we must be careful on destruction of the scopes sharing them.
type Scope struct {
	id                  ora.ScopeID
	factories           map[ora.ComponentFactoryId]ComponentFactory
	allocatedComponents map[ora.Ptr]allocatedComponent
	lifetime            time.Duration
	endOfLifeAt         atomic.Pointer[time.Time]
	channel             AtomicRef[Channel]
	chanDestructor      AtomicRef[func()]
	destroyed           AtomicRef[bool]
	eventLoop           *EventLoop
	lastMessageType     ora.EventType
	ctx                 context.Context
	cancelCtx           func()
	sessionID           SessionID
	tempDirMutex        sync.Mutex
	tempRootDir         string
	tempDir             string
	nextFileSeqNo       int64
}

func NewScope(ctx context.Context, tempRootDir string, id ora.ScopeID, lifetime time.Duration, factories map[ora.ComponentFactoryId]ComponentFactory) *Scope {

	scopeCtx, cancel := context.WithCancel(ctx)
	s := &Scope{
		id:                  id,
		factories:           factories,
		allocatedComponents: map[ora.Ptr]allocatedComponent{},
		lifetime:            lifetime,
		eventLoop:           NewEventLoop(),
		ctx:                 scopeCtx,
		cancelCtx:           cancel,
		tempRootDir:         tempRootDir,
	}

	s.eventLoop.SetOnPanicHandler(func(p any) {
		s.Publish(ora.ErrorOccurred{
			Type:    ora.ErrorOccurredT,
			Message: fmt.Sprintf("panic in event loop: %v", p),
		})
	})
	s.channel.Store(NopChannel{})
	s.Tick()

	return s
}

// OnStreamReceive provides a side channel for sending large streams of blobs which must not
// block the lightweight and responsive connected event Channel.
// If e.g. this scope is part of a http web server, each uploaded file must trigger a call here.
// The stream is consumed immediately and stored within a temporary file synchronously with a metadata sidecar file.
// The so-stored file is kept as long as this scope is alive and can be inspected any time.
// Afterward, the component is invoked over the application event looper.
// Note, that the origin could also be from different sources, like the content resolver within an Android App
// directly issued over FFI calls.
func (s *Scope) OnStreamReceive(stream StreamReader) (e error) {
	s.Tick()

	fileId := atomic.AddInt64(&s.nextFileSeqNo, 1)
	tmpDir, err := s.getTempDir()
	if err != nil {
		return err
	}

	absPath := filepath.Join(tmpDir, fmt.Sprintf("%d.tmp", fileId))
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open tmp file for write: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && e == nil {
			e = err
		}
	}()

	size, err := io.Copy(file, stream)
	if err != nil {
		return fmt.Errorf("cannot copy data: %w", err)
	}

	// hash that file, often of interest at the domain level e.g. for deduplication or content-addressed-storage
	hashStr, err := s.readHash(absPath)
	if err != nil {
		return fmt.Errorf("cannot calculate hash: %w", err)
	}

	// create sidecar file
	meta := tmpFileInfo{
		AbsolutePath: absPath,
		FName:        stream.Name(),
		FSize:        size,
		Hash:         hashStr,
		CreatedAt:    time.Now(),
		SeqNum:       fileId,
		Scope:        stream.ScopeID(),
		Receiver:     stream.Receiver(),
	}

	sidecarBuf, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("cannot marshal meta: %w", err)
	}

	sidecarFile := filepath.Join(tmpDir, fmt.Sprintf("%d.json", fileId))
	if err := os.WriteFile(sidecarFile, sidecarBuf, 0600); err != nil {
		return fmt.Errorf("cannot write sidecar file: %w", err)
	}

	s.eventLoop.Post(func() {
		var receiver Component
		if cmp, ok := s.allocatedComponents[meta.Receiver]; ok {
			receiver = cmp.Component
		} else {
			for _, holder := range s.allocatedComponents {
				if cmp, ok := holder.RenderState.elements[meta.Receiver]; ok {
					receiver = cmp
				}
			}
		}

		if receiver == nil {
			slog.Error("receiver component for data stream not found")
			return
		}

		switch receiver := receiver.(type) {
		case FileReceiver:
			f, err := os.Open(meta.AbsolutePath)
			if err != nil {
				slog.Error("cannot open file for reading in looper: %w", err)
				return
			}

			receiver.OnFileReceived(newTmpFile(meta, f))
		default:
			slog.Error("receiver component for data stream has no compatible receiver interface", "type", fmt.Sprintf("%T", receiver))
		}
	})

	return nil
}

func (s *Scope) readHash(path string) (string, error) {
	hasher := sha512.New512_256()
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file: %w", err)
	}

	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *Scope) getTempDir() (string, error) {
	s.tempDirMutex.Lock()
	defer s.tempDirMutex.Unlock()

	if s.tempDir != "" {
		return s.tempDir, nil
	}

	// we don't know where the temp root is. It may be in our apps home (e.g. in shared hosting environments)
	// or in the systems temp dir.
	path := filepath.Join(s.tempRootDir, string(s.id))
	//0600 means that only the owner can read and write
	if err := os.MkdirAll(path, 0600); err != nil {
		return "", fmt.Errorf("cannot create temp dir for scope: %s: %w", path, err)
	}

	slog.Info("created temp dir for scope", "path", path)

	s.tempDir = path
	return path, nil
}

// Connect attaches the given channel to this Scope immediately. There must be exact 1 Scope per Channel.
// The use case is, that Scopes can be transferred from one channel to another easily.
func (s *Scope) Connect(c Channel) {
	if c == nil {
		c = NopChannel{}
	}

	slog.Info("scope connected to channel", slog.String("scopeId", string(s.id)), slog.String("channel", fmt.Sprintf("%T", c)))

	s.chanDestructor.With(func(destructor func()) func() {
		s.channel.Store(c)

		if destructor != nil {
			destructor()
		}

		return c.Subscribe(func(msg []byte) error {
			defer s.eventLoop.Tick()
			return s.handleMessage(msg)
		})
	})
}

func (s *Scope) handleMessage(buf []byte) error {
	s.Tick()

	t, err := ora.Unmarshal(buf)
	if err != nil {
		return err
	}

	s.eventLoop.Post(func() {
		s.handleEvent(t)
		// todo handleEvent may have caused already a rendering. Should we omit to avoid sending multiple times?
		if s.lastMessageType != ora.ComponentInvalidatedT {
			s.renderIfRequired()
		}
	})

	return nil
}

func (s *Scope) Publish(evt ora.Event) {
	switch evt := evt.(type) {
	case ora.ComponentInvalidated:
		s.lastMessageType = evt.Type
	default:
		s.lastMessageType = ""
	}

	if err := s.channel.Load().Publish(ora.Marshal(evt)); err != nil {
		slog.Error(err.Error())
	}
}

// Tick marks this scope as used and moves the EOL forward.
func (s *Scope) Tick() {
	eol := time.Now().Add(s.lifetime)
	s.endOfLifeAt.Store(&eol)
}

// EOL returns the current estimated end of life.
func (s *Scope) EOL() time.Time {
	return *s.endOfLifeAt.Load()
}

// sendAck eventually sends an acknowledged message, if id is not 0. This is intentional, so a sender
// can optimize for performance.
func (s *Scope) sendAck(id ora.RequestId) {
	if id == 0 {
		return
	}

	s.Publish(ora.Acknowledged{
		Type:      ora.AcknowledgedT,
		RequestId: id,
	})
}

func (s *Scope) sendPing() {
	s.Publish(ora.Ping{
		Type: ora.PingT,
	})
}

// only for event loop
func (s *Scope) renderIfRequired() {
	for _, component := range s.allocatedComponents {
		if IsDirty(component.Component) {
			//slog.Info("component is dirty", slog.Int("ptr", int(component.Component.ID())))
			s.Publish(s.render(0, component.Component))
		}
	}
}

// only for event loop
func (s *Scope) render(requestId ora.RequestId, component Component) ora.ComponentInvalidated {
	Freeze(component)
	defer Unfreeze(component)

	rs := s.allocatedComponents[component.ID()].RenderState
	rs.Clear()
	rs.Scan(component)
	ClearDirty(component)

	return ora.ComponentInvalidated{
		Type:      ora.ComponentInvalidatedT,
		RequestId: requestId,
		Component: component.Render(),
	}
}

// Destroy frees all allocated components and removes factory pointers.
// The scope is of no use afterward.
// Do never call this from the event loop.
func (s *Scope) Destroy() {
	s.destroyed.With(func(destroyed bool) bool {
		if destroyed {
			return true
		}

		s.eventLoop.Post(func() {
			s.destroy()
		})

		s.eventLoop.Destroy()

		return true
	})

}

// only for event loop
func (s *Scope) destroy() {
	s.cancelCtx()
	for _, component := range s.allocatedComponents {
		invokeDestructors(component)
	}

	s.channel.Store(NewPrintChannel()) // detach
	clear(s.allocatedComponents)
	//	clear(s.factories) // clearing this map would cause a data race, even though we use the factory as read-only
}

// only for event loop
func (s *Scope) handleSessionAssigned(evt ora.SessionAssigned) {
	s.sessionID = SessionID(evt.SessionID)
}

// only for event loop
func invokeDestructors(component allocatedComponent) {
	component.Window.viewRoot.Destroy()

	if closer, ok := component.Component.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			slog.Error("error on closing Component", slog.Any("err", err), slog.String("type", fmt.Sprintf("%T", component)))
		}
	}

	if destroyer, ok := component.Component.(Destroyable); ok {
		destroyer.Destroy()
	}
}
