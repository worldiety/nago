package ui

import (
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
)

type History struct {
	p *Page
}

func (h *History) Back() {
	h.p.sendMsg(messageHistoryBack{Type: "HistoryBack"})
}

func (h *History) Open(pageId PageID, params Values) {
	h.p.sendMsg(messageHistoryPushState{
		Type:   "HistoryPushState",
		PageID: string(pageId),
		State:  params,
	})
}

func (h *History) OpenURL(url string, target string) {
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
