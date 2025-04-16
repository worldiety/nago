// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

type Components struct {
	Schemas         map[string]*Schema         `json:"schemas"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes"`
}
