// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

type DataTypeFormat string

const (
	DTInt32    DataTypeFormat = "int32"
	DTInt64    DataTypeFormat = "int64"
	DTFloat32  DataTypeFormat = "float"
	DTFloat64  DataTypeFormat = "double"
	DTPassword DataTypeFormat = "password"
)
