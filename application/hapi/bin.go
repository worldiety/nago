package hapi

import (
	"go.wdy.de/nago/pkg/oas/v31"
	"io"
	"log/slog"
	"net/http"
)

// ToBinary supports the following annotations:
// - required:"true"
func ToBinary[In any](fn func(in In) (io.Reader, error)) ResponseOption[In] {
	return func(doc *oas.OpenAPI, r *ResponseBuilder[In]) {
		r.contentType = "application/octet-stream"
		r.schema = schemaOf[io.ReadCloser](doc)
		r.handler = func(in In, writer http.ResponseWriter, request *http.Request) {
			reader, err := fn(in)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				slog.Error("failed to handle request", "error", err.Error())
				return
			}

			defer func() {
				if outReadCloser, ok := reader.(io.ReadCloser); ok {
					if err := outReadCloser.Close(); err != nil {
						slog.Error("failed to close binary response reader", "err", err.Error())
					}
				}
			}()

			if _, err := io.Copy(writer, reader); err != nil {
				slog.Error("failed to write response", "error", err.Error())
				return
			}
		}
	}
}
