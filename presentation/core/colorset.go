// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"fmt"
	"regexp"

	"go.wdy.de/nago/presentation/proto"
)

// A ColorSet marks a simple struct with public color fields (like [Colors]) to be a set of colors.
// It returns its unique namespace and has a Default behavior, as a fallback.
// Even though this looks quite cumbersome, for just defining some custom colors, it will play out its strength,
// when designing custom views with complex color sets. If a component requires 10 additional color values and
// you combine 10 different components, you already have to manage and define 100 unstructured color values
// at configuration time. Therefore, we have namespaces and the type safety.
type ColorSet interface {
	// Default returns an initialized color set of the same type as self but with sensible default values set.
	Default(scheme ColorScheme) ColorSet
	// Namespace must be unique within an entire application. "ora" is reserved.
	Namespace() NamespaceName
}

type NamespaceName string

var validColorNamespaceNameRegex = regexp.MustCompile(`ˆ[A-Za-z0-9_\-]+$`)

func (c NamespaceName) Valid() bool {
	return validColorNamespaceNameRegex.MatchString(string(c))
}

type ColorScheme uint

func (c ColorScheme) String() string {
	switch c {
	case Light:
		return "light"
	case Dark:
		return "dark"
	case System:
		return "system"
	default:
		return fmt.Sprintf("%d", c)
	}
}

const (
	System ColorScheme = ColorScheme(proto.System)
	Light  ColorScheme = ColorScheme(proto.Light)
	Dark   ColorScheme = ColorScheme(proto.Dark)
)
