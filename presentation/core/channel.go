// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import "go.wdy.de/nago/presentation/proto"

// A Channel abstracts a concrete technical implementation about the communication between frontend and backend.
// An implementation must not deadlock while publishing messages from the subscription callback.
// Considered implementation priorities are:
//   - websocket
//   - foreign function interface call (e.g. a go lib embedded into a native mobile application).
//   - server side events
//   - mqtt
type Channel interface {
	// Subscribe receives messages from the channel.
	Subscribe(f func(msg []byte) error) (destroy func())
	// Publish writes the message into the channel.
	Publish(msg []byte) error
}

// An EventChannel is an optional extension of a [Channel] for transports within the same process, like
// the nagotest package. If the channel connected to a [Scope] implements it, the Scope publishes its events
// directly without encoding them into the binary protocol. Incoming events can be injected
// using [Scope.Dispatch].
type EventChannel interface {
	Channel
	// PublishEvent receives an outgoing event of the Scope. The event must be treated as read-only.
	PublishEvent(evt proto.NagoEvent) error
}
