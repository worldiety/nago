// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package proto declares the messages between the arena program in a browser and its backend. Both import it, the
// program as part of the subset and the backend as normal Go, so they can not disagree about a message.
package proto

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Move reports the position of a player to the backend.
type Move struct {
	ID   string
	X, Z float64
	Seq  int32
}

// Hit tells a player, that it was hit, without a request.
type Hit struct {
	Target string
	Damage int32
	Seq    int32
}

// Ack confirms a hit to the backend.
type Ack struct {
	ID  string
	Seq int32
	HP  int32
}

// Pos is the position of a player.
type Pos struct {
	ID   string
	X, Z float64
}

// World is the state of all players, which the backend sends from time to time.
type World struct {
	Players list.List[Pos]
}
