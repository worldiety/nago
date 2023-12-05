package ui

import (
	"encoding/json"
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
