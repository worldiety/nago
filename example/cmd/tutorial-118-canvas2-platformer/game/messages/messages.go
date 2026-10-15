// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package messages declares the messages between the game in the browser and its backend.
package messages

// Hello asks the backend for the highscore, when the game starts.
type Hello struct {
	Player string
}

// Finished reports a completed level.
type Finished struct {
	Player  string
	Seconds float64
	Coins   int32
}

// Highscore is the best time of all players.
type Highscore struct {
	Holder  string
	Seconds float64
}
