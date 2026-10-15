// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

// Animation is an animation of a view, e.g. AnimatePulse. See TStack.Animation.
type Animation uint32

const (
	AnimateNone       Animation = 0
	AnimateTransition Animation = 1
	AnimatePulse      Animation = 2
	AnimateBounce     Animation = 3
	AnimatePing       Animation = 4
	AnimateSpin       Animation = 5
)
