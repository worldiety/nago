package application

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"net/http"
)

type rawEndpoint struct {
	pattern string
	handler http.HandlerFunc
}

type Resource interface {
	configureResource(c *Configurator) ora.URI
}

type Bytes []byte

func (r Bytes) configureResource(c *Configurator) ora.URI {
	token := string(ora.NewScopeID())
	pattern := fmt.Sprintf("/api/ora/v1/resource/%s", token)
	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: func(writer http.ResponseWriter, request *http.Request) {
			if _, err := writer.Write(r); err != nil {
				slog.Error("failed to write response", "err", err)
			}
		},
	})

	return ora.URI(pattern)
}

// Resource registers the given resource. It will likely result in an additional endpoint which looks like
// /api/ora/v1/resource/<some random identifier>
func (c *Configurator) Resource(r Resource) ora.URI {
	return r.configureResource(c)
}
