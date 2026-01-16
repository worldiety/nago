// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

type SMTP struct {
	Name          string `value:"Mein SMTP Server"`
	Host          string
	Port          int `value:"587"`
	Username      string
	Password      string `style:"secret"`
	SenderAddress string `value:"" label:"Absenderadresse" supportingText:"Wenn leer, wird der Username verwendet, ansonsten hat diese Absenderadresse Vorrang."`
	_             string `credentialName:"SMTP Postausgangsserver" credentialDescription:"Ein Postausgangsserver wird benötigt, um E-Mails zu verschicken." credentialLogo:"https://www.thunderbird.net/media/img/thunderbird/favicon-196.png"`
}

func (SMTP) Credentials() bool {
	return true
}

func (s SMTP) GetName() string {
	return s.Name
}

func (s SMTP) IsZero() bool {
	return s == SMTP{}
}
