// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

type WorkflowStatus int

const (
	Pending WorkflowStatus = iota
	Error
	Finished
)

type workflowState struct {
	ID     ID             `json:"id"`
	Status WorkflowStatus `json:"status"`
}

type Repository interface {
}
