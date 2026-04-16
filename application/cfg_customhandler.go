// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"net/http"

	http2 "go.wdy.de/nago/presentation/core/http"
)

// HandleFunc allows for Nago instances to inject a http handler.
// The given handler will respond to any http method and only one can be registered.
// See also [Configurator.HandleMethod] to only register a handler for a specific method.
// Note, that this is not possible on non-server platforms like mobile applications.
func (c *Configurator) HandleFunc(pattern string, handler http.HandlerFunc) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: handler,
	})
}

// HandleMethod allows for Nago instances to inject a http handler which ever responds to the given http method.
// See also [Configurator.HandleFunc].
func (c *Configurator) HandleMethod(method string, pattern string, handler http.HandlerFunc) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		method:  method,
		pattern: pattern,
		handler: handler,
	})
}

// HandleFuncSubject allows for Nago instances to inject a http handler which responds to the given http method.
// The handler will be wrapped by a session and user middleware.
// See also [Configurator.HandleFunc] and [Configurator.HandleMethod].
func (c *Configurator) HandleFuncSubject(pattern string, handler http2.SubjectHandlerFunc) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	modSessions, err := c.SessionManagement()
	if err != nil {
		return err
	}

	modUsers, err := c.UserManagement()
	if err != nil {
		return err
	}

	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: http2.NewSubjectHandlerFunc(
			modSessions.UseCases.FindUserSessionByID,
			modUsers.UseCases.SubjectFromUser,
			modUsers.UseCases.GetAnonUser,
			handler,
		),
	})

	return nil
}
