// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"iter"
	"reflect"
	"strings"

	"go.wdy.de/nago/application/user"
)

type Instance string

// InstanceKey is just an id which is a composite of <workflow-id>/<instance-id> to optimize queries to find all instances
// for a specific workflow. We intentionally do not place any state inside a struct, because that would cause
// a lot of trouble in a distributed system, e.g. if we would put an instance-local storage into an instance struct:
// Each action is executed concurrently and in a possible future, we would need a distributed mutex to update it
// safely without loosing unaffected key-value pairs.
type InstanceKey string

func NewInstanceKey(workflow ID, instance Instance) InstanceKey {
	return InstanceKey(workflow) + "/" + InstanceKey(instance)
}

func (i InstanceKey) Split() (ID, Instance, bool) {
	tokens := strings.Split(string(i), "/")
	if len(tokens) != 2 {
		return "", "", false
	}

	return ID(tokens[0]), Instance(tokens[1]), true
}

// Typename consists of <Package>.<Type>
type Typename string

func NewTypename(t reflect.Type) Typename {
	return Typename(t.PkgPath() + "." + t.Name())
}

type FindInstances func(subject user.Subject, id ID) iter.Seq2[Instance, error]
type CreateInstance func(subject user.Subject, id ID) (Instance, error)
