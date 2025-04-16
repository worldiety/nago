package hapi

import (
	"bytes"
	"github.com/worldiety/enum/json"
	"go.wdy.de/nago/pkg/oas/v31"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
)

// ToJSON supports the following annotations:
// - required:"true"
func ToJSON[In, Out any](fn func(in In) (Out, error)) ResponseOption[In] {
	return func(doc *oas.OpenAPI, r *ResponseBuilder[In]) {
		r.contentType = "application/json; charset=utf-8"
		r.schema = schemaOf[Out](doc)
		r.handler = func(in In, writer http.ResponseWriter, request *http.Request) {
			out, err := fn(in)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				slog.Error("failed to handle request", "error", err.Error())
				return
			}

			buf, err := json.Marshal(out)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				slog.Error("failed to encode json response", "error", err.Error())
				return
			}

			if _, err := writer.Write(buf); err != nil {
				slog.Error("failed to write response", "error", err.Error())
				return
			}
		}
	}
}

func JSONFromBody[In, Model any](fn func(dst *In, model Model) error) RequestOption[In] {
	return func(doc *oas.OpenAPI, b *RequestBuilder[In]) {
		b.handlers = append(b.handlers, requestSchema[In]{
			contentType: "application/json; charset=utf-8",
			schema:      schemaOf[Model](doc),
			fieldName:   "",
			intoModel: func(dst *In, w http.ResponseWriter, r *http.Request) error {
				mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil {
					slog.Error("failed to parse content type", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if mediaType != "application/json" {
					slog.Error("unsupported content type", "contentType", mediaType)
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				dec := json.NewDecoder(r.Body)
				defer r.Body.Close()

				// decision: I cannot remember a single case, where a newer client is sending newer additional fields to
				// a legacy server and expecting it not to fail. The newer client must check the server version and must
				// send a backwards compatible request.
				dec.DisallowUnknownFields()
				var model Model
				if err := dec.Decode(&model); err != nil {
					slog.Error("failed to decode JSON", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if err := fn(dst, model); err != nil {
					slog.Error("failed to handle request", "error", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				return nil
			},
		})

	}
}

func FilesFromFormField[In any](fieldname string, fn func(dst *In, files []*multipart.FileHeader) error) RequestOption[In] {
	return func(doc *oas.OpenAPI, b *RequestBuilder[In]) {
		b.handlers = append(b.handlers, requestSchema[In]{
			contentType: "multipart/form-data",
			schema:      schemaOf[[]multipart.File](doc),
			fieldName:   fieldname,
			intoModel: func(dst *In, w http.ResponseWriter, r *http.Request) error {
				mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil {
					slog.Error("failed to parse content type", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if mediaType != "multipart/form-data" {
					slog.Error("unsupported content type", "contentType", mediaType)
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				// 1MiB in memory, rest on disk
				if err := r.ParseMultipartForm(1024 * 1024); err != nil {
					slog.Error("failed to parse multipart form", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if r.MultipartForm == nil {
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				values := r.MultipartForm.File[fieldname]
				return fn(dst, values)
			},
		})
	}
}

func JSONFromFormField[In, Model any](fieldname string, fn func(dst *In, model Model) error) RequestOption[In] {
	return func(doc *oas.OpenAPI, b *RequestBuilder[In]) {
		b.handlers = append(b.handlers, requestSchema[In]{
			contentType: "multipart/form-data",
			schema:      schemaOf[Model](doc),
			fieldName:   fieldname,
			intoModel: func(dst *In, w http.ResponseWriter, r *http.Request) error {
				mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil {
					slog.Error("failed to parse content type", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if mediaType != "multipart/form-data" {
					slog.Error("unsupported content type", "contentType", mediaType)
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				// 1MiB in memory, rest on disk
				if err := r.ParseMultipartForm(1024 * 1024); err != nil {
					slog.Error("failed to parse multipart form", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if r.MultipartForm == nil {
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				values := r.MultipartForm.Value[fieldname]
				if len(values) == 0 {
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				dec := json.NewDecoder(bytes.NewReader([]byte(values[0])))
				defer r.Body.Close()

				// decision: I cannot remember a single case, where a newer client is sending newer additional fields to
				// a legacy server and expecting it not to fail. The newer client must check the server version and must
				// send a backwards compatible request.
				dec.DisallowUnknownFields()
				var model Model
				if err := dec.Decode(&model); err != nil {
					slog.Error("failed to decode JSON from form", "err", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				if err := fn(dst, model); err != nil {
					slog.Error("failed to handle request", "error", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return errorAlreadyHandled
				}

				return nil
			},
		})

	}
}
