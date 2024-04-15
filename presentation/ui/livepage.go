package ui

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

const TextMessage = 1

type SessionID = string

// TODO I don't like that, because we cannot popup dialogs from nowwhere, but perhaps that is a good thing?
type ModalOwner interface {
	Modals() *SharedList[core.Component]
}

// deprecated we want probably the scope instead or any other interface
type Wire interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	// Values contains those values, which have been passed from the callers, e.g. intent parameters or url query
	// parameters. This depends on the actual frontend.
	Values() Values
	// User is never nil. Check [auth.User.Valid]. You must not keep the User instance over a long time, because
	// it will change over time, either due to refreshing tokens or because the user is logged out.
	User() auth.User
	// Context returns the wire-lifetime context. Contains additional injected types like User or Logger.
	Context() context.Context
	// Remote information, which is especially useful for audit logs.
	Remote() Remote

	// ClientSession is a unique identifier, which is assigned to a client using a cookie mechanism. This is a
	// pure random string and belongs to a distinct client instance. It is shared across multiple pages on the client,
	// especially when using multiple tabs or browser windows.
	ClientSession() SessionID
}

// deprecated what is this for?
type Remote interface {
	// Addr denotes the physical remote layer. This is useless behind a proxy.
	Addr() string
	// ForwardedFor interprets different http headers. This works only behind a trusted proxy,
	// because it is prone to spoofing.
	ForwardedFor() string
}

type PageInstanceToken string

type Page struct {
	id         CID
	wire       Wire
	body       *Shared[core.Component]
	modals     *SharedList[core.Component]
	history    *History
	properties []core.Property
	onDestroy  []func()
}

func NewPage(w Wire, with func(page *Page)) *Page {
	p := &Page{wire: w, id: nextPtr()}
	p.history = &History{p: p}
	p.body = NewShared[core.Component]("body")
	p.modals = NewSharedList[core.Component]("modals")
	//p.token = NewShared[string]("token")
	//p.token.Set(nextToken())
	p.properties = []core.Property{p.body, p.modals}
	//p.maxMemory = 1024
	//p.renderState = core.NewRenderState()
	if with != nil {
		with(p)
	}
	return p
}

func (p *Page) ID() CID {
	return p.id
}

func (p *Page) Type() protocol.ComponentType {
	return protocol.PageT
}

func (p *Page) Properties(yield func(core.Property) bool) {
	for _, property := range p.properties {
		if !yield(property) {
			return
		}
	}
}

func (p *Page) Render() protocol.Component {
	return protocol.Page{
		Ptr:    p.id,
		Type:   protocol.PageT,
		Body:   renderComponentProp(p.body, p.body),
		Modals: renderComponentsProp(p.modals, p.modals),
	}
}

func (p *Page) Body() *Shared[LiveComponent] {
	return p.body
}

func (p *Page) Token() PageInstanceToken {
	//return PageInstanceToken(p.token.Get())
	panic("implement me")
}

func (p *Page) Modals() *SharedList[core.Component] {
	return p.modals
}

// deprecated: the scope[component-id] must be invalidated instead
func (p *Page) Invalidate() {
	logging.FromContext(p.wire.Context()).Info("page invalidated: re-render")

	/*p.renderState.Clear()
	// TODO make also a real component
	var tmp []jsonComponent
	p.modals.Each(func(component LiveComponent) {
		tmp = append(tmp, marshalComponent(p.renderState, component))
	})
	/*
		p.sendMsg(messageFullInvalidate{
			Type:   "Invalidation",
			Root:   marshalComponent(p.renderState, p.body.Get()),
			Modals: tmp,
			Token:  string(p.Token()),
		})*/
}

// deprecated: this does not belong to a page, but the applications scope
func (p *Page) History() *History {
	return p.history
}

// HandleHTTP provides classic http inter-operation with this page. This is required e.g. for file uploads
// using multipart forms etc.
// deprecated: entirely the wrong place for this
func (p *Page) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageToken := r.Header.Get("x-page-token")
	if pageToken == "" {
		pageToken = query.Get("page")
	}

	//if pageToken != p.token.Get() {
	//	w.WriteHeader(http.StatusNotFound)
	//	return
	//}
	// TODO where and how to handle that???
	/*
		switch r.URL.Path {
		case "/api/v1/upload":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
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
		case "/api/v1/download":
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			downloadToken := DownloadToken(r.Header.Get("x-download-token"))
			if downloadToken == "" {
				downloadToken = DownloadToken(query.Get("download"))
			}

			opener := p.renderState.downloads[downloadToken]
			if opener == nil {
				logging.FromContext(r.Context()).Warn("download request received but have no handler", slog.String("download-token", string(downloadToken)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			reader, err := opener()
			if err != nil {
				logging.FromContext(r.Context()).Warn("download request received but cannot open stream", slog.String("download-token", string(downloadToken)), slog.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(w, reader); err != nil {
				logging.FromContext(r.Context()).Warn("download request received but cannot complete data transfer", slog.String("download-token", string(downloadToken)), slog.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

	*/
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
	/*	_, buf, err := p.wire.ReadMessage()
		if err != nil {
			slog.Default().Error("failed to receive ws message", slog.Any("err", err))
			return err
		}

		var batch msgBatch
		if err := json.Unmarshal(buf, &batch); err != nil {
			slog.Default().Error("cannot decode ws batch message", slog.Any("err", err))
			return err
		}

		if len(batch.Messages) == 0 {
			slog.Default().Error("received empty message batch from client, it should not do that")
			return nil
		}

		for _, buf := range batch.Messages {
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

			case "updateJWT":
				// nothing to do, this is handled transparently by the wire layer itself because the encoding
				// and auth details are implementation dependent

			default:
				slog.Default().Error("protocol not implemented: " + m.Type)
			}
		}

		if IsDirty(p.body.Get()) || p.body.Dirty() || p.modals.Dirty() {
			p.body.SetDirty(false)
			p.modals.SetDirty(false)
			SetDirty(p.body.Get(), false)
			p.Invalidate()
		}
	*/
	return nil
}

func (p *Page) Close() error {
	slog.Default().Info("live page is dead")

	for _, f := range p.onDestroy {
		f()
	}

	return nil
}

func (p *Page) AddOnDestroy(f func()) {
	p.onDestroy = append(p.onDestroy, f)
}

/*
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

type msgBatch struct {
	Messages []json.RawMessage `json:"tx"`
}

type callFunc struct {
	ID CID `json:"id"`
}

type setProperty struct {
	ID    CID `json:"id"`
	Value any `json:"value"`
}


func callIt(rs *internal.RenderState, call callFunc) {
	f := rs.funcs[call.ID]
	if !f.Nil() {
		slog.Default().Info(fmt.Sprintf("func called %d", f.ID()))
		f.Invoke()
	}
}

func setProp(rs *internal.RenderState, set setProperty) {
	v := rs.props[set.ID]
	if v != nil {
		if err := v.setValue(fmt.Sprintf("%v", set.Value)); err != nil {
			slog.Default().Error(fmt.Sprintf("cannot set property %d = %v, reason: %v", v.ID(), v.Unwrap(), err))
		}
		slog.Default().Info(fmt.Sprintf("value set %d = %v", v.ID(), v.Unwrap()))
	}
}
*/
