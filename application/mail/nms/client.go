// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package nms contains the client of the Nago Mail Service (NMS), an HTTP mail relay for Nago instances.
//
// A Nago instance gets its credentials in one of two ways:
//   - a refresh token configured by the operator, see [Options.Token].
//   - a token exchange: the instance asks the service for a token. Either it proves, that it is reachable under its
//     public https origin, because the service calls the origin back and expects the nonce of the exchange, see
//     [Nonces]. This works for origins the service accepts, typically the wildcard subdomains of a hub. Or it calls
//     from an address the service knows, e.g. a developer on localhost from the office network. The service then
//     admits it by its address and calls nothing back. Which applies is decided by the service alone.
//
// The refresh token is exchanged for short-lived access tokens, which authorize sending. The service may revoke
// any token at any time, the [Service] then enrolls again.
package nms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.wdy.de/nago/pkg/xhttp"
)

// DefaultEndpoint is used, if no endpoint has been configured.
const DefaultEndpoint = "https://mail.worldiety.nago.app"

// Error codes of the service.
const (
	CodeUnauthorized     = "unauthorized"
	CodeTokenRevoked     = "token_revoked"
	CodeRefreshInvalid   = "refresh_invalid"
	CodeOriginNotAllowed = "origin_not_allowed"
	CodeOriginUnverified = "origin_unverified"
	CodeTenantExists     = "tenant_exists"
	CodeRateLimited      = "rate_limited"

	CodeInvalidSender      = "invalid_sender"
	CodeSenderNotAllowed   = "sender_not_allowed"
	CodeInvalidRecipient   = "invalid_recipient"
	CodeMissingBody        = "missing_body"
	CodeAttachmentTooLarge = "attachment_too_large"
	CodeHeaderNotAllowed   = "header_not_allowed"
	CodeDuplicate          = "duplicate"
)

// Message status values of a send result.
const (
	StatusQueued = "queued"
	StatusError  = "error"
)

type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type Attachment struct {
	Filename      string `json:"filename"`
	ContentType   string `json:"contentType"`
	ContentID     string `json:"contentId,omitempty"` // only for inlined attachments
	Base64Content string `json:"base64Content"`
}

type Message struct {
	From               Address           `json:"from"`
	To                 []Address         `json:"to,omitempty"`
	Cc                 []Address         `json:"cc,omitempty"`
	Bcc                []Address         `json:"bcc,omitempty"`
	ReplyTo            *Address          `json:"replyTo,omitempty"`
	Subject            string            `json:"subject"`
	TextPart           string            `json:"textPart,omitempty"`
	HTMLPart           string            `json:"htmlPart,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Attachments        []Attachment      `json:"attachments,omitempty"`
	InlinedAttachments []Attachment      `json:"inlinedAttachments,omitempty"`
	// CustomID identifies the message for the caller. The service sends a message with a known CustomID only once
	// within 24 hours and answers with the known result instead.
	CustomID string `json:"customId,omitempty"`
}

type MessageError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

type MessageResult struct {
	Status    string         `json:"status"`
	CustomID  string         `json:"customId,omitempty"`
	MessageID string         `json:"messageId,omitempty"`
	Errors    []MessageError `json:"errors,omitempty"`
}

type Limits struct {
	PerHour int `json:"perHour,omitempty"`
	PerDay  int `json:"perDay,omitempty"`
}

// Enrollment is the result of a token exchange.
type Enrollment struct {
	Tenant        string   `json:"tenant"`
	Refresh       string   `json:"refresh"`
	SenderDomains []string `json:"senderDomains"`
	Limits        Limits   `json:"limits"`
}

// Tokens is the result of a token refresh. The refresh token is rotated.
type Tokens struct {
	Access    string `json:"access"`
	ExpiresIn int    `json:"expiresIn"` // seconds
	Refresh   string `json:"refresh"`
}

type Health struct {
	Status        string   `json:"status"`
	Tenant        string   `json:"tenant,omitempty"`
	SenderDomains []string `json:"senderDomains,omitempty"`
	Revoked       bool     `json:"revoked,omitempty"`
	Limits        struct {
		PerHour        int `json:"perHour,omitempty"`
		PerDay         int `json:"perDay,omitempty"`
		RemainingToday int `json:"remainingToday,omitempty"`
	} `json:"limits"`
}

// Error is a failed request, which the service answered with an error status.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("nago mail service: status %d: %s", e.Status, e.Message)
	}

	return fmt.Sprintf("nago mail service: status %d: %s: %s", e.Status, e.Code, e.Message)
}

// Revoked is true, if the used token is no longer valid and a new one must be obtained.
func (e *Error) Revoked() bool {
	return e.Code == CodeTokenRevoked || e.Code == CodeRefreshInvalid
}

// Client talks to the service. It does not retry, the caller decides.
type Client struct {
	Endpoint string
	HTTP     *http.Client // optional
	Timeout  time.Duration
}

func (c Client) request(ctx context.Context, path string) *xhttp.Request {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	req := xhttp.NewRequest().
		Context(ctx).
		URL(strings.TrimRight(c.Endpoint, "/") + path).
		Timeout(timeout).
		Assert2xx(true).
		ToLimit(1024 * 1024)

	if c.HTTP != nil {
		req = req.Client(c.HTTP)
	}

	return req
}

// Exchange asks the service for a refresh token of the given origin. The service calls the origin back and
// expects the nonce, see [Nonces].
func (c Client) Exchange(ctx context.Context, origin, nonce string) (Enrollment, error) {
	var res Enrollment
	err := c.request(ctx, "/v1/exchange").
		BodyJSON(map[string]string{"origin": origin, "nonce": nonce}).
		ToJSON(&res).
		Post()
	if err != nil {
		return Enrollment{}, asError(err)
	}

	if res.Refresh == "" {
		return Enrollment{}, fmt.Errorf("nago mail service: exchange response has no refresh token")
	}

	return res, nil
}

// Token exchanges the refresh token for an access token and a new refresh token.
func (c Client) Token(ctx context.Context, refresh string) (Tokens, error) {
	var res Tokens
	err := c.request(ctx, "/v1/token").
		BodyJSON(map[string]string{"refresh": refresh}).
		ToJSON(&res).
		Post()
	if err != nil {
		return Tokens{}, asError(err)
	}

	if res.Access == "" {
		return Tokens{}, fmt.Errorf("nago mail service: token response has no access token")
	}

	return res, nil
}

// Send hands the messages over to the service. A request accepted as a whole returns one result per message,
// even if single messages have been rejected.
func (c Client) Send(ctx context.Context, access string, msgs ...Message) ([]MessageResult, error) {
	var res struct {
		Messages []MessageResult `json:"messages"`
	}

	err := c.request(ctx, "/v1/send").
		BearerAuthentication(access).
		BodyJSON(map[string]any{"messages": msgs}).
		ToJSON(&res).
		Post()
	if err != nil {
		return nil, asError(err)
	}

	if len(res.Messages) != len(msgs) {
		return nil, fmt.Errorf("nago mail service: expected %d results but got %d", len(msgs), len(res.Messages))
	}

	return res.Messages, nil
}

// Health returns the state of the service and, if an access token is given, of the tenant.
func (c Client) Health(ctx context.Context, access string) (Health, error) {
	var res Health
	req := c.request(ctx, "/v1/health").ToJSON(&res)
	if access != "" {
		req = req.BearerAuthentication(access)
	}

	if err := req.Get(); err != nil {
		return Health{}, asError(err)
	}

	return res, nil
}

// asError turns an error status into an [Error] and keeps any other error, e.g. of the connection.
func asError(err error) error {
	var status xhttp.UnexpectedStatusCodeError
	if !errors.As(err, &status) {
		return err
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	res := &Error{Status: status.StatusCode}
	if json.Unmarshal(status.Body, &body) == nil && body.Error.Code != "" {
		res.Code = body.Error.Code
		res.Message = body.Error.Message
	} else {
		res.Message = http.StatusText(status.StatusCode)
	}

	return res
}
