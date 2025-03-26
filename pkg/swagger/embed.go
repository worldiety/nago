package swagger

import (
	"embed"
	"encoding/json"
	"go.wdy.de/nago/pkg/oas/v30"
	"log/slog"
	"net/http"
)

//go:embed api/doc
var files embed.FS

func HandleOAS(spec *oas.OpenAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		buf, err := json.Marshal(spec)
		if err != nil {
			slog.Error("failed to marshal oas spec: %v", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(buf); err != nil {
			slog.Error("failed to write oas spec: %v", "err", err)
			return
		}
	}
}
