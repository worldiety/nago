package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
)

const TextMessage = 1

type Wire interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Values() Values
}

type LivePage struct {
	wire    Wire
	body    LiveComponent
	history *History
}

func NewLivePage(w Wire, with func(page *LivePage)) *LivePage {
	p := &LivePage{wire: w}
	p.history = &History{p: p}
	if with != nil {
		with(p)
	}
	return p
}

func (p *LivePage) SetBody(c LiveComponent) {
	p.body = c
}

func (p *LivePage) Invalidate() {
	p.sendMsg(messageFullInvalidate{
		Type: "Invalidation",
		Root: marshalComponent(p.body),
	})
}

func (p *LivePage) History() *History {
	return p.history
}

func (p *LivePage) HandleMessage() error {
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
		callIt(p.body, call)
	case "setProp":
		var call setProperty
		if err := json.Unmarshal(buf, &call); err != nil {
			panic(fmt.Errorf("cannot happen: %w", err))
		}

		setProp(p.body, call)

	default:
		slog.Default().Error("protocol not implemented: " + m.Type)
	}

	if IsDirty(p.body) {
		SetDirty(p.body, false)
		p.Invalidate()
	}

	return nil
}

func (p *LivePage) Close() error {
	slog.Default().Info("live page is dead")
	return nil
}

func (p *LivePage) sendMsg(t any) {
	buf, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Errorf("implementation failure: %w", err))
	}
	if err := p.wire.WriteMessage(TextMessage, buf); err != nil {
		slog.Default().Error("failed to write websocket message", slog.Any("err", err))
	}
}

type messageFullInvalidate struct {
	Type string        `json:"type"` // value=Invalidation
	Root jsonComponent `json:"root"`
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
	dst.Functions().Each(func(idx int, f *Func) {
		if f.ID() == call.ID {
			f.Invoke()
			slog.Default().Info(fmt.Sprintf("func called %d", f.ID()))
		}
	})

	dst.Children().Each(func(idx int, component LiveComponent) {
		callIt(component, call)
	})
}

func setProp(dst LiveComponent, set setProperty) {
	if dst == nil {
		return
	}
	dst.Properties().Each(func(idx int, v Property) {
		if v.ID() == set.ID {
			if err := v.SetValue(fmt.Sprintf("%v", set.Value)); err != nil {
				slog.Default().Error(fmt.Sprintf("cannot set property %d = %v, reason: %v", v.ID(), v.Value(), err))
			}
			slog.Default().Info(fmt.Sprintf("value set %d = %v", v.ID(), v.Value()))
		}
	})

	dst.Children().Each(func(idx int, component LiveComponent) {
		setProp(component, set)
	})
}

type History struct {
	p *LivePage
}

func (h *History) Back() {
	h.p.sendMsg(messageHistoryBack{Type: "HistoryBack"})
}

func (h *History) PushForward(pageId PageID, params Values) {
	h.p.sendMsg(messageHistoryPushState{
		Type:   "HistoryPushState",
		PageID: string(pageId),
		State:  params,
	})
}

func (h *History) Open(url string, target string) {
	h.p.sendMsg(messageHistoryOpen{
		Type:   "HistoryOpen",
		URL:    url,
		Target: target,
	})
}

type Values map[string]string

func UnmarshalValues[Dst any](src Values) (Dst, error) {
	var params Dst
	t := reflect.TypeOf(params)
	v := reflect.ValueOf(&params).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if n, ok := f.Tag.Lookup("name"); ok {
			name = n
		}

		value, ok := src[name]
		if !ok {
			continue
		}

		switch f.Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(value)
		case reflect.Int:
			x, err := strconv.Atoi(value)
			if err != nil {
				slog.Default().Error("cannot parse integer path variable", slog.Any("err", err))
			}

			v.Field(i).SetInt(int64(x))
		case reflect.Int64:
			x, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				slog.Default().Error("cannot parse integer path variable", slog.Any("err", err))
			}

			v.Field(i).SetInt(x)

		case reflect.Uint64:
			x, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				slog.Default().Error("cannot parse integer path variable", slog.Any("err", err))
			}

			v.Field(i).SetUint(x)
		default:
			return params, fmt.Errorf("cannot parse '%s' into %T.%s", value, params, f.Name)
		}

	}

	return params, nil
}
