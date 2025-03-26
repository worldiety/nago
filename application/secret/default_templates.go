// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

type SMTP struct {
	Name     string `value:"Mein SMTP Server"`
	Host     string
	Port     int `value:"587"`
	Username string
	Password string `style:"secret"`
	_        string `credentialName:"SMTP Postausgangsserver" credentialDescription:"Ein Postausgangsserver wird benötigt, um E-Mails zu verschicken." credentialLogo:"https://www.thunderbird.net/media/img/thunderbird/favicon-196.png"`
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

type Jira struct {
	Name  string `value:"Meine Jira Instanz"`
	EMail string
	Token string `style:"secret"`
	_     string `credentialName:"Jira API" credentialDescription:"E-Mail und Token zur API Anbindung einer Jira Cloud Instanz definieren." credentialLogo:"https://wac-cdn.atlassian.com/assets/img/favicons/atlassian/mstile-144x144.png"`
}

func (Jira) Credentials() bool {
	return true
}

func (s Jira) GetName() string {
	return s.Name
}

func (s Jira) IsZero() bool {
	return s == Jira{}
}

type BookStack struct {
	Name          string `value:"Meine BookStack Instanz"`
	URL           string
	TokenID       string
	TokenPassword string
	Check         bool
	_             string `credentialName:"BookStack" credentialDescription:"URL, Token-ID und Token-Passwort zur API Anbindung einer Bookstack Instanz definieren." credentialLogo:"https://www.bookstackapp.com/images/favicon-196x196.png"`
}

func (BookStack) Credentials() bool {
	return true
}

func (s BookStack) GetName() string {
	return s.Name
}

func (s BookStack) IsZero() bool {
	return s == BookStack{}
}
