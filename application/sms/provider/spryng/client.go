// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package spryng

import (
	"net/http"
	"time"

	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xstrings"
)

type SMS struct {
	Body       string   `json:"body"`
	Encoding   string   `json:"encoding"`
	Route      string   `json:"route"`
	Originator string   `json:"originator"`
	Recipients []string `json:"recipients"` // Empfängernummer im MSISDN Format (E.164-Format)
}

type SMSResponse struct {
	Id          string    `json:"id"`
	Encoding    string    `json:"encoding"`
	Originator  string    `json:"originator"`
	Body        string    `json:"body"`
	Reference   string    `json:"reference"`
	Credits     float64   `json:"credits"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CanceledAt  string    `json:"canceled_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Links       struct {
		Self string `json:"self"`
	} `json:"links"`
}

type Client struct {
	token string
	cl    *http.Client
	group *xhttp.RequestGroup
	base  string
}

func NewClient(settings Settings) *Client {
	return &Client{
		token: settings.Token,
		cl: &http.Client{
			Timeout: time.Second * 30,
		},
		group: xhttp.NewRequestGroup().RateLimit(settings.RPS),
		base:  "https://rest.spryngsms.com/v1",
	}
}

func (c *Client) Send(sms message.SendRequested) (SMSResponse, error) {
	msg := SMS{
		Body:       sms.Body,
		Encoding:   "auto",
		Route:      "business",
		Originator: xstrings.EllipsisEnd(string(sms.Originator), 11),
		Recipients: []string{sms.Recipient.String()},
	}

	var resp SMSResponse
	err := xhttp.NewRequest().
		Group(c.group).
		BaseURL(c.base).
		URL("messages").
		Assert2xx(true).
		BearerAuthentication(c.token).
		BodyJSON(msg).
		ToJSON(&resp).
		Post()

	if err != nil {
		return resp, err
	}

	return resp, nil
}
