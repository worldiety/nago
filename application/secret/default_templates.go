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
	Password      string   `style:"secret"`
	SenderAddress string   `value:"" label:"Absenderadresse" supportingText:"Muss eine gültige E-Mail-Adresse sein. Wenn leer oder ungültig, wird der Username verwendet, sofern dieser eine gültige E-Mail-Adresse ist."`
	// RateLimitPerHour limits the amount of send attempts per hour. 0 means unlimited.
	RateLimitPerHour int `label:"Max. Mails pro Stunde" supportingText:"Optionale Ratenbegrenzung des Anbieters. 0 bedeutet unbegrenzt."`
	// RateLimitPerDay limits the amount of send attempts within 24 hours. 0 means unlimited.
	RateLimitPerDay int `label:"Max. Mails pro Tag" supportingText:"Optionale Ratenbegrenzung des Anbieters, z.B. 500. 0 bedeutet unbegrenzt."`
	_             struct{} `credentialName:"SMTP Postausgangsserver" credentialDescription:"Ein Postausgangsserver wird benötigt, um E-Mails zu verschicken." credentialLogo:"https://www.thunderbird.net/media/img/thunderbird/favicon-196.png"`
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
