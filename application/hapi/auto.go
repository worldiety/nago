// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"go.wdy.de/nago/pkg/oas/v30"
	"net/http"
)

// Auto inspects the given T using reflection to read or write in various formats. Supported fields are
// defined as follows.
//
// # status
//   - if a field 'Status int' is found, the status will be set dynamically. Note that this cannot be rendered in
//     the OpenAPI documentation.
//
// # header
//   - supported field types: int, string, time
//   - field tag 'header' to specify the header name
//   - timeFormat in Go time.Format to specify the format of the time to parse from
//
// # body
//   - inspect the fields and pick Body if available
//   - structs are encoded and decoded as JSON
//   - a []byte slice will be written as is
//   - everything else is not supported
//
// At any time, you can implement your custom format including OpenAPI definitions by conforming to the
// [Input] and [Output] interfaces.
type Auto[T any] struct {
	Body T
}

func (a *Auto[T]) Write(w http.ResponseWriter, t *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (a *Auto[T]) DescribeOutput(op *oas.Operation) {
	//TODO implement me
	panic("implement me")
}

func (a *Auto[T]) Read(w http.ResponseWriter, t *http.Request) error {
	//TODO implement me
	panic("implement me")
}

func (a *Auto[T]) DescribeInput(op *oas.Operation) {
	//TODO implement me
	panic("implement me")
}
