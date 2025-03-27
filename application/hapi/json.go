package hapi

import (
	"github.com/worldiety/enum/json"
	"go.wdy.de/nago/pkg/oas/v30"
	"log/slog"
	"mime"
	"net/http"
)

type None struct{}

func (h *None) DescribeOutput(op *oas.Operation) {
}

func (h *None) Read(w http.ResponseWriter, t *http.Request) error {
	return nil
}

func (h *None) DescribeInput(op *oas.Operation) {
}

func (h *None) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type JSON[T any] struct {
	Body T
}

func (h *JSON[T]) Read(w http.ResponseWriter, r *http.Request) error {
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
	if err := dec.Decode(&h.Body); err != nil {
		slog.Error("failed to decode JSON", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return errorAlreadyHandled
	}

	return nil
}

func (h *JSON[T]) Write(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	buf, err := json.Marshal(h.Body)
	if err != nil {
		return err
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}

func (h *JSON[T]) DescribeOutput(op *oas.Operation) {
	if op.Responses == nil {
		op.Responses = oas.Responses{}
	}

	var zero T

	op.Responses["200"] = &oas.Response{
		Headers: nil,
		Content: map[string]oas.MediaType{
			"200": oas.MediaType{
				Schema:  oas.Schema{Ref: "#/components/schemas/Nago"},
				Example: zero,
				Encoding: map[oas.MediaTypeRange]oas.Encoding{
					"application/json": oas.Encoding{},
				},
			},
		},
	}
}

func (h *JSON[T]) DescribeInput(op *oas.Operation) {}
