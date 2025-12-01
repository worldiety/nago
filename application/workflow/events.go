// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"reflect"

	"go.wdy.de/nago/pkg/xjson"
)

func init() {
	xjson.RegisterSelf(reflect.TypeFor[ActionInvoked]())
	xjson.RegisterSelf(reflect.TypeFor[ActionCompletedSuccessfully]())
	xjson.RegisterSelf(reflect.TypeFor[InstanceCreated]())
	xjson.RegisterSelf(reflect.TypeFor[InstanceEventEnvelope]())
	xjson.RegisterSelf(reflect.TypeFor[InstanceStopped]())
}

type ActionInvoked struct {
	Instance Instance
	Action   Typename
}

type ActionCompletedSuccessfully struct {
	Workflow ID
	Instance Instance
	Action   Typename
}

type ActionFailed struct {
	Workflow  ID
	Instance  Instance
	Action    Typename
	Error     string
	ErrorType Typename
}

type InstanceCreated struct {
	Workflow ID
	Instance Instance
}

type InstanceEventEnvelope struct {
	Workflow ID
	Instance Instance
	Event    any
}

type InstanceStopped struct {
	Instance Instance
}
