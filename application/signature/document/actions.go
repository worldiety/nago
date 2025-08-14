// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"context"

	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/workflow"
)

type SendMails struct {
	sendMail workflow.Publisher[mail.SendMailRequested]
}

func (a *SendMails) Configure(cfg *workflow.Configuration) error {
	a.sendMail = workflow.GlobalEvent[mail.SendMailRequested](cfg)
	cfg.SetName("E-Mails an Unterzeichner versenden")
	cfg.SetDescription(`SendMails is triggered by an external [SignaturesRequested] event 
and sends to each signatory an email to place his sign 
on the document. The mail will contain a link to the 
according page. If a recipient signs it, the page will 
trigger a [SignatureCaptured] event.`)
	return nil
}

func (a *SendMails) OnEvent(ctx context.Context, evt SignaturesRequested) error {
	return nil
}

type UpdateDocumentState struct {
}

func (a *UpdateDocumentState) Configure(cfg *workflow.Configuration) error {
	cfg.SetName("Dokumentenstatus aktualisieren")
	workflow.LocalEvent[SignatoriesCompleted](cfg)
	workflow.LocalEvent[SignaturesRequested](cfg)
	cfg.SetDescription("If all signatures have been collected, the [SignatoriesCompleted]\nevent is raised and the workflow is completed.")
	return nil
}

func (a *UpdateDocumentState) OnEvent(ctx context.Context, evt SignatureCaptured) error {
	return nil
}
