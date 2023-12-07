package ui

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/container/slice"
	"log/slog"
)

const TextMessage = 1

type Wire interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Values() Values
}

type Page struct {
	id         CID
	wire       Wire
	body       *Shared[LiveComponent]
	modals     *SharedList[LiveComponent]
	history    *History
	properties slice.Slice[Property]
}

func NewPage(w Wire, with func(page *Page)) *Page {
	p := &Page{wire: w, id: nextPtr()}
	p.history = &History{p: p}
	p.body = NewShared[LiveComponent]("body")
	p.modals = NewSharedList[LiveComponent]("modals")
	p.properties = slice.Of[Property](p.body, p.modals)
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

func (p *Page) Modals() *SharedList[LiveComponent] {
	return p.modals
}

func (p *Page) Invalidate() {
	// TODO make also a real component
	var tmp []jsonComponent
	p.modals.Each(func(component LiveComponent) {
		tmp = append(tmp, marshalComponent(component))
	})
	p.sendMsg(messageFullInvalidate{
		Type:   "Invalidation",
		Root:   marshalComponent(p.body.Get()),
		Modals: tmp,
	})
}

func (p *Page) History() *History {
	return p.history
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
		callIt(p, call)
	case "setProp":
		var call setProperty
		if err := json.Unmarshal(buf, &call); err != nil {
			panic(fmt.Errorf("cannot happen: %w", err))
		}

		setProp(p, call)

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

func callIt(dst LiveComponent, call callFunc) {
	if dst == nil {
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
	})
}

func setProp(dst LiveComponent, set setProperty) {
	if dst == nil {
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
	})
}
