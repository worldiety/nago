package application

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"net/http"
	"time"
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
	sum := sha256.Sum256(r)
	token := hex.EncodeToString(sum[:])
	pattern := fmt.Sprintf("/api/ora/v1/resource/%s", token)
	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: func(writer http.ResponseWriter, request *http.Request) {
			// enable aggressive caching, because we have a stable resource identifier based on a hash sum
			writer.Header().Set("Cache-Control", "public, max-age=31536000")
			expires := time.Now().Add(365 * 24 * time.Hour)
			writer.Header().Set("Expires", expires.Format(http.TimeFormat))

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
