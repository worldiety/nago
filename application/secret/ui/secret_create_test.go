// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"reflect"
	"testing"
)

type visibleCredentials struct {
	_ struct{} `credentialName:"Visible"`
}

type hiddenCredentials struct {
	_ struct{} `credentialName:"Hidden" credentialHidden:"true"`
}

func TestCredentialTypeSpecHidden(t *testing.T) {
	if newCredentialTypeSpec(reflect.TypeFor[visibleCredentials]()).hidden {
		t.Fatal("visible credentials must not be hidden")
	}

	if !newCredentialTypeSpec(reflect.TypeFor[hiddenCredentials]()).hidden {
		t.Fatal("hidden credentials must be hidden")
	}
}
