package application

import (
	"net/http"
)

// HandleFunc allows for Nago instances to inject a http handler.
// Note, that this is not possible on non-server platforms like mobile applications.
func (c *Configurator) HandleFunc(pattern string, handler http.HandlerFunc) {
	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: handler,
	})
}
