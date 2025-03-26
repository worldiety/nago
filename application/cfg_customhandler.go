package application

import (
	"net/http"
)

// HandleFunc allows for Nago instances to inject a http handler.
// The given handler will respond to any http method and only one can be registered.
// See also [Configurator.HandleMethod] to only register a handler for a specific method.
// Note, that this is not possible on non-server platforms like mobile applications.
func (c *Configurator) HandleFunc(pattern string, handler http.HandlerFunc) {
	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: handler,
	})
}

// HandleMethod allows for Nago instances to inject a http handler which ever responds to the given http method.
// See also [Configurator.HandleFunc].
func (c *Configurator) HandleMethod(method string, pattern string, handler http.HandlerFunc) {
	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		method:  method,
		pattern: pattern,
		handler: handler,
	})
}
