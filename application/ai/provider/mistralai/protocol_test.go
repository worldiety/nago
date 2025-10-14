// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("BpC2pIoLxJoiL8Xd2GsvH97nhGqlg0mQ") // TODO nicht einchecken
	/*t.Log(c.CreateConversion(CreateConversionRequest{
		Inputs: "test unterhaltung",
		Model:  "mistral-large-latest",
	}))*/
	// conv_0199c4939ae274ee98bd8ee438df0ac2 / conv_0199c928567f7483869a8c4dc4b53bce

	res, err := c.CreateAgent(CreateAgentRequest{
		Name:  "Test-Agent",
		Model: "mistral-large-latest",
	})
	t.Logf("%+v: %v", res, err)
}
