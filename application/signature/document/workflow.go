// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/workflow"
	"reflect"
)

const Workflow = "nago.signature.workflow"

func NewWorkflow() workflow.DeclareOptions {
	return workflow.DeclareOptions{
		ID:          Workflow,
		Name:        "Dokument unterzeichnen",
		Description: "Workflow zur Unterzeichnung eines Dokumentes durch mehrere Unterzeichner",
		Actions: []workflow.Action{
			workflow.Start[SignaturesRequested]{},
			&UpdateDocumentState{},
			&SendMails{},
			workflow.Stop[SignatoriesCompleted]{},
		},
		Transitions: []workflow.Transition{
			{
				Type:     workflow.HumanInTheLoop,
				InEvent:  reflect.TypeFor[mail.SendMailRequested](),
				OutEvent: reflect.TypeFor[SignatureCaptured](),
			},
		},
	}
}
