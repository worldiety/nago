// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package target is an entry point, whose last result is an error, which the aspect of the test denies with a value
// or with an abort.
package target

import (
	"go.wdy.de/nago/pkg/gostrict/internal/whole/n10/guard"
	"go.wdy.de/nago/pkg/gostrict/internal/whole/store"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// Info is the first result of Main: a basic value, a number and a list, whose zero values n10/pass passes on to the
// host.
type Info struct {
	Name string
	N    int32
	Tags list.List[string]
}

// Main records its run under target/<id>; with reject, guard.Reject ends it before it records after/<id>.
func Main(id string, reject bool) (Info, error) {
	store.Put("target/"+id, "ran")
	if reject {
		guard.Reject("rejected " + id)
		store.Put("after/"+id, "ran")
	}
	return Info{Name: "ran", N: 7, Tags: list.Of("x")}, nil
}
