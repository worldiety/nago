package ui

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/logging"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

const TextMessage = 1

type Wire interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Values() Values
}

type PageInstanceToken string

type Page struct {
	id          CID
	wire        Wire
	body        *Shared[LiveComponent]
	modals      *SharedList[LiveComponent]
	history     *History
	properties  slice.Slice[Property]
	renderState *renderState
	token       String
	maxMemory   int64
}

func NewPage(w Wire, with func(page *Page)) *Page {
	p := &Page{wire: w, id: nextPtr()}
	p.history = &History{p: p}
	p.body = NewShared[LiveComponent]("body")
	p.modals = NewSharedList[LiveComponent]("modals")
	p.token = NewShared[string]("token")
	p.token.Set(nextToken())
	p.properties = slice.Of[Property](p.body, p.modals, p.token)
	p.maxMemory = 1024
	p.renderState = newRenderState()
	if with != nil {
		with(p)
	}
	return p
}

func (p *Page) ID() CID {
	return p.id
}

func (p *Page) Type() string {
	return "Page"
}

func (p *Page) Properties() slice.Slice[Property] {
	return p.properties
}

func (p *Page) Body() *Shared[LiveComponent] {
	return p.body
}

func (p *Page) Token() PageInstanceToken {
	return PageInstanceToken(p.token.Get())
}

func (p *Page) Modals() *SharedList[LiveComponent] {
	return p.modals
}

func (p *Page) Invalidate() {
	p.renderState.Clear()
	// TODO make also a real component
	var tmp []jsonComponent
	p.modals.Each(func(component LiveComponent) {
		tmp = append(tmp, marshalComponent(p.renderState, component))
	})
	p.sendMsg(messageFullInvalidate{
		Type:   "Invalidation",
		Root:   marshalComponent(p.renderState, p.body.Get()),
		Modals: tmp,
		Token:  string(p.Token()),
	})
}

func (p *Page) History() *History {
	return p.history
}

// HandleHTTP provides classic http inter-operation with this page. This is required e.g. for file uploads
// using multipart forms etc.
func (p *Page) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	pageToken := r.Header.Get("x-page-token")
	if pageToken != p.token.Get() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.URL.Path {
	case "/api/v1/upload":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		uploadToken := UploadToken(r.Header.Get("x-upload-token"))
		handler := p.renderState.uploads[uploadToken]
		if handler == nil || handler.onUploadReceived == nil {
			logging.FromContext(r.Context()).Warn("upload received but have no handler", slog.String("upload-token", string(uploadToken)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := r.ParseMultipartForm(p.maxMemory); err != nil {
			logging.FromContext(r.Context()).Warn("cannot parse multipart form", slog.Any("err", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var files []FileUpload
		for _, headers := range r.MultipartForm.File {
			for _, header := range headers {
				files = append(files, httpMultipartFile{header: header})
			}
		}

		handler.onUploadReceived(files)
		p.Invalidate() // TODO race condition?!?!
	}
}

type httpMultipartFile struct {
	header *multipart.FileHeader
}

func (h httpMultipartFile) Size() int64 {
	return h.header.Size
}

func (h httpMultipartFile) Name() string {
	return h.header.Filename
}

func (h httpMultipartFile) Open() (io.ReadSeekCloser, error) {
	return h.header.Open()
}

func (h httpMultipartFile) Sys() any {
	return h.header
}

func (p *Page) HandleMessage() error {
	_, buf, err := p.wire.ReadMessage()
	if err != nil {
		slog.Default().Error("failed to receive ws message", slog.Any("err", err))
		return err
	}

	fmt.Println("got message", string(buf))
	var m msg
	if err := json.Unmarshal(buf, &m); err != nil {
		slog.Default().Error("cannot decode ws message", slog.Any("err", err))
		return err
	}

	switch m.Type {
	case "callFn":
		var call callFunc
		if err := json.Unmarshal(buf, &call); err != nil {
			panic(fmt.Errorf("cannot happen: %w", err))
		}
		callIt(p.renderState, call)
	case "setProp":
		var call setProperty
		if err := json.Unmarshal(buf, &call); err != nil {
			panic(fmt.Errorf("cannot happen: %w", err))
		}

		setProp(p.renderState, call)

	default:
		slog.Default().Error("protocol not implemented: " + m.Type)
	}

	if IsDirty(p.body.Get()) || p.body.Dirty() || p.modals.Dirty() {
		p.body.SetDirty(false)
		p.modals.SetDirty(false)
		SetDirty(p.body.Get(), false)
		p.Invalidate()
	}

	return nil
}

func (p *Page) Close() error {
	slog.Default().Info("live page is dead")
	return nil
}

func (p *Page) sendMsg(t any) {
	buf, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Errorf("implementation failure: %w", err))
	}
	if err := p.wire.WriteMessage(TextMessage, buf); err != nil {
		slog.Default().Error("failed to write websocket message", slog.Any("err", err))
	}
}

type messageFullInvalidate struct {
	Type   string          `json:"type"` // value=Invalidation
	Root   jsonComponent   `json:"root"`
	Modals []jsonComponent `json:"modals"`
	Token  string          `json:"token"`
}

type messageHistoryBack struct {
	Type string `json:"type"` // value=HistoryBack
}

type messageHistoryPushState struct {
	Type   string            `json:"type"` // value=HistoryPushState
	PageID string            `json:"pageId"`
	State  map[string]string `json:"state"`
}

type messageHistoryOpen struct {
	Type   string `json:"type"` // value=HistoryOpen
	URL    string `json:"url"`
	Target string `json:"target"`
}

type msg struct {
	Type string `json:"type"`
}

type callFunc struct {
	ID CID `json:"id"`
}

type setProperty struct {
	ID    CID `json:"id"`
	Value any `json:"value"`
}

func callIt(rs *renderState, call callFunc) {
	/*if dst == nil {
		return
	}
	Functions(dst, func(f *Func) {
		if f.ID() == call.ID && !f.Nil() {
			f.Invoke()
			slog.Default().Info(fmt.Sprintf("func called %d", f.ID()))
		}
	})

	Children(dst, func(c LiveComponent) {
		callIt(c, call)
	})*/
	f := rs.funcs[call.ID]
	if !f.Nil() {
		slog.Default().Info(fmt.Sprintf("func called %d", f.ID()))
		f.Invoke()
	}
}

func setProp(rs *renderState, set setProperty) {
	/*if dst == nil {
		return
	}
	dst.Properties().Each(func(idx int, v Property) {
		if v.ID() == set.ID {
			if err := v.setValue(fmt.Sprintf("%v", set.Value)); err != nil {
				slog.Default().Error(fmt.Sprintf("cannot set property %d = %v, reason: %v", v.ID(), v.value(), err))
			}
			slog.Default().Info(fmt.Sprintf("value set %d = %v", v.ID(), v.value()))
		}
	})

	Children(dst, func(c LiveComponent) {
		setProp(c, set)
	})*/
	v := rs.props[set.ID]
	if v != nil {
		if err := v.setValue(fmt.Sprintf("%v", set.Value)); err != nil {
			slog.Default().Error(fmt.Sprintf("cannot set property %d = %v, reason: %v", v.ID(), v.value(), err))
		}
		slog.Default().Info(fmt.Sprintf("value set %d = %v", v.ID(), v.value()))
	}
}
