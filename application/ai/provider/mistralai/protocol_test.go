// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	token, ok := os.LookupEnv("MISTRAL_API_TOKEN")
	if !ok {
		t.Log("MISTRAL_API_TOKEN environment variable not set")
		return
	}

	c := NewClient(token)
	/*t.Log(c.CreateConversion(CreateConversionRequest{
		Inputs: "test unterhaltung",
		Model:  "mistral-large-latest",
	}))*/
	// conv_0199c4939ae274ee98bd8ee438df0ac2 / conv_0199c928567f7483869a8c4dc4b53bce

	/*	res, err := c.CreateAgent(CreateAgentRequest{
			Name:  "Test-Agent",
			Model: "mistral-large-latest",
		})
		t.Logf("%+v: %v", res, err)*/

	for agent, err := range c.ListAgents() {
		if err != nil {
			t.Error(err)
		}

		slog.Info(agent.Id)
	}
}
