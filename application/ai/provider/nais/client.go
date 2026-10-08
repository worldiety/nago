// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package nais contains the provider of the Nago AI Service (NAIS), a gateway to several AI providers, which counts
// the tokens every instance uses.
//
// The service speaks the Messages API of Anthropic, whatever provider a model lies with, so the provider is the
// Anthropic provider pointed at the service. Only the credentials differ: a Nago instance enrolls by itself, like with
// the Nago Mail Service, and needs no secret in the vault:
//   - a refresh token configured by the operator, see [Options.Token], or
//   - a token exchange: either the instance proves that it is reachable under its public https origin, because the
//     service calls the origin back and expects the nonce of the exchange, see [Nonces], or it calls from an address
//     the service knows, e.g. a developer on localhost from the office network.
//
// The refresh token is exchanged for short-lived access tokens, which authorize the calls. The service may revoke
// any token at any time, the [Service] then enrolls again.
//
// The provider is only compiled in and started, if the application side-imports this package:
//
//	import _ "go.wdy.de/nago/application/ai/provider/nais"
package nais

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
const DefaultEndpoint = "https://ai.worldiety.nago.app"

// Error codes of the service.
const (
	CodeUnauthorized     = "unauthorized"
	CodeTokenRevoked     = "token_revoked"
	CodeRefreshInvalid   = "refresh_invalid"
	CodeOriginNotAllowed = "origin_not_allowed"
	CodeOriginUnverified = "origin_unverified"
	CodeRateLimited      = "rate_limited"
	CodeBudgetExhausted  = "budget_exhausted"
	CodeModelNotAllowed  = "model_not_allowed"
)

// Model is a model the tenant of the instance may call.
type Model struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Limits are the limits of the tenant.
type Limits struct {
	TokensPerDay int `json:"tokensPerDay,omitempty"`
}

// Enrollment is the result of a token exchange.
type Enrollment struct {
	Tenant       string  `json:"tenant"`
	Refresh      string  `json:"refresh"`
	Models       []Model `json:"models"`
	DefaultModel string  `json:"defaultModel"`
	Limits       Limits  `json:"limits"`
}

// Tokens is the result of a token refresh. The refresh token is rotated.
type Tokens struct {
	Access    string `json:"access"`
	ExpiresIn int    `json:"expiresIn"` // seconds
	Refresh   string `json:"refresh"`
}

// Health is the state of the service and, with an access token, of the tenant.
type Health struct {
	Status       string  `json:"status"`
	Tenant       string  `json:"tenant,omitempty"`
	Models       []Model `json:"models,omitempty"`
	DefaultModel string  `json:"defaultModel,omitempty"`
	Limits       struct {
		TokensPerDay   int  `json:"tokensPerDay,omitempty"`
		RemainingToday *int `json:"remainingToday,omitempty"`
	} `json:"limits"`
}

// Error is a failed request, which the service answered with an error status. The token handling answers
// {"error":{"code","message"}}, the Messages API answers its own format with the same code added; both are read.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("nago ai service: status %d: %s", e.Status, e.Message)
	}

	return fmt.Sprintf("nago ai service: status %d: %s: %s", e.Status, e.Code, e.Message)
}

// Revoked is true, if the used token is no longer valid and a new one must be obtained by an exchange.
func (e *Error) Revoked() bool {
	return e.Code == CodeTokenRevoked || e.Code == CodeRefreshInvalid
}

// Client talks to the token handling of the service. It does not retry, the caller decides.
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

// Exchange asks the service for a refresh token of the given origin. The service calls the origin back and expects
// the nonce, see [Nonces].
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
		return Enrollment{}, fmt.Errorf("nago ai service: exchange response has no refresh token")
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
		return Tokens{}, fmt.Errorf("nago ai service: token response has no access token")
	}

	return res, nil
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

	return errorOf(status.StatusCode, status.Body)
}

// errorOf reads an error answer of the service.
func errorOf(status int, body []byte) *Error {
	var e struct {
		Error struct {
			Code    string `json:"code"`
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	res := &Error{Status: status}
	if json.Unmarshal(body, &e) == nil && (e.Error.Code != "" || e.Error.Message != "") {
		res.Code = e.Error.Code
		res.Message = e.Error.Message
	} else {
		res.Message = http.StatusText(status)
	}

	return res
}
