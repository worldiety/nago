// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow_test

import (
	"context"
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/workflow"
	"go.wdy.de/nago/application/workflow/instance"
	"testing"
)

type ProofOfEducationSubmitted struct {
	DocID   string
	Trainer string
	Trainee string
}

type FirstOfMonthReached struct {
}

type TrainerSigned struct {
	ID workflow.IID
}

type TraineeSigned struct {
	ID workflow.IID
}

type DocumentCompleted struct {
}

type SignatureReceived struct {
	ID  workflow.IID
	UID user.ID
}

type AllSigned struct {
	workflow.IID
}

type SendMailRequired struct {
	ID    instance.ID
	Who   string
	What  string
	DocID string
}

type MailNotSend struct {
	ID instance.ID
}

type SendMailIfReceived struct {
	sendMails workflow.Publisher[SendMailRequired]
	mailErr   workflow.Publisher[MailNotSend]
}

func (w *SendMailIfReceived) Configure(cfg *workflow.Configuration) error {
	w.sendMails = workflow.MayPublish[SendMailRequired](cfg)
	w.mailErr = workflow.MayPublish[MailNotSend](cfg)
	cfg.SetName("Some Businesslogic to decide what to do")
	cfg.SetDescription("A lot is going on here.")
	return nil
}

func (w *SendMailIfReceived) OnEvent(ctx context.Context, evt SignatureReceived) error {
	w.sendMails(SendMailRequired{
		ID:    "",
		Who:   "",
		What:  "",
		DocID: "",
	})
	return nil
}

func TestNewWorkflow4(t *testing.T) {
	uc := workflow.NewUseCases()
	wid := option.Must(uc.Declare(user.SU(), workflow.DeclareOptions{
		ID:          "1234",
		Name:        "Aubina Workflow",
		Description: "The Aubina Workflow requires a proof read of the document and the signatures of the trainer and a trainee.",
		Actions: []workflow.Action{
			workflow.Start[SignatureReceived]{},
			&SendMailIfReceived{},
			workflow.Stop[SendMailRequired]{},
			workflow.Stop[MailNotSend]{},
		},
	}))

	fmt.Println(string(option.Must(uc.Render(user.SU(), wid))))

	option.MustZero(uc.DispatchEvent(user.SU(), SignatureReceived{
		ID:  "1234",
		UID: "",
	}))

}
