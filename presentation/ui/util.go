package ui

import (
	"encoding/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
	"log/slog"
	"net/http"
)

func writeJson(w http.ResponseWriter, r *http.Request, v any) {
	w.Header().Set("content-type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		// TODO grab from request
		slog.Default().Error("failed to encode and write response", slog.Any("err", err))
	}
}

func must2[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func renderComponentProp(property core.Property, p core.Iterable[core.Component]) protocol.Property[protocol.Component] {
	var first core.Component
	p.Iter(func(component core.Component) bool {
		first = component
		return false
	})

	return protocol.Property[protocol.Component]{
		Ptr:   property.ID(),
		Value: first.Render(),
	}
}

func renderComponentsProp(property core.Property, p core.Iterable[core.Component]) protocol.Property[[]protocol.Component] {
	var res []protocol.Component
	p.Iter(func(component core.Component) bool {
		res = append(res, component.Render())
		return true
	})

	return protocol.Property[[]protocol.Component]{
		Ptr:   property.ID(),
		Value: res,
	}
}
