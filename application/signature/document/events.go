// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"go.wdy.de/nago/application/user"
)

type Signatory struct {
	Firstname string
	Lastname  string
	User      user.ID // Note that the user id may be empty and is just a hint
	Email     user.Email
}

type SignaturesRequested struct {
	_           any           `label:"Unterzeichnung angefordert"`
	Resource    user.Resource // the resource to be signed
	Signatories []Signatory
}

type SignatoriesCompleted struct {
	Resource    user.Resource
	Signatories []Signatory
}

type SignatureCaptured struct {
	Resource  user.Resource
	Signatory Signatory
}
