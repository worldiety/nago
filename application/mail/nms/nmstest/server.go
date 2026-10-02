// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package nmstest provides an in-memory Nago Mail Service for tests.
package nmstest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/mail/nms"
)

// Server implements the API of the service. The call back of an exchange is delegated to [Server.VerifyOrigin],
// which by default accepts every origin without calling it.
type Server struct {
	URL string

	// VerifyOrigin checks the nonce of an exchange. Default accepts everything.
	VerifyOrigin func(origin, nonce string) bool
	// SenderDomains derives the sender domains of an enrolled origin. Default is the host of the origin.
	SenderDomains func(origin string) []string
	// AccessLifetime of issued access tokens. Default is 1 hour.
	AccessLifetime time.Duration

	mutex     sync.Mutex
	refresh   map[string]string // token -> tenant
	access    map[string]string // token -> tenant
	tenants   map[string][]string
	sent      []nms.Message
	exchanges int
	tokens    int
	sends     int
}

// NewServer starts a server, which is closed at the end of the test.
func NewServer(t testing.TB) *Server {
	s := &Server{
		refresh: map[string]string{},
		access:  map[string]string{},
		tenants: map[string][]string{},
	}

	srv := httptest.NewServer(s)
	t.Cleanup(srv.Close)
	s.URL = srv.URL
	return s
}

// IssueRefresh creates a tenant with the given sender domains and returns its refresh token, like an operator does.
func (s *Server) IssueRefresh(tenant string, senderDomains ...string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	token := "nmr_" + randHex()
	s.refresh[token] = tenant
	s.tenants[tenant] = senderDomains
	return token
}

// RevokeAll revokes every token of every tenant.
func (s *Server) RevokeAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.refresh = map[string]string{}
	s.access = map[string]string{}
}

// Sent returns all accepted messages.
func (s *Server) Sent() []nms.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return append([]nms.Message(nil), s.sent...)
}

// Calls returns the amount of exchange, token and send requests.
func (s *Server) Calls() (exchanges, tokens, sends int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.exchanges, s.tokens, s.sends
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/exchange":
		s.handleExchange(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/token":
		s.handleToken(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/send":
		s.handleSend(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/health":
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	default:
		writeError(w, http.StatusNotFound, "not_found", "unknown endpoint")
	}
}

func (s *Server) handleExchange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Origin string `json:"origin"`
		Nonce  string `json:"nonce"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	s.mutex.Lock()
	s.exchanges++
	verify := s.VerifyOrigin
	domains := s.SenderDomains
	s.mutex.Unlock()

	if verify != nil && !verify(req.Origin, req.Nonce) {
		writeError(w, http.StatusForbidden, nms.CodeOriginUnverified, "the origin did not answer with the nonce")
		return
	}

	tenant := strings.TrimPrefix(req.Origin, "https://")
	senderDomains := []string{tenant}
	if domains != nil {
		senderDomains = domains(req.Origin)
	}

	s.mutex.Lock()
	// a new exchange revokes the former tokens of the tenant
	for token, t := range s.refresh {
		if t == tenant {
			delete(s.refresh, token)
		}
	}
	s.mutex.Unlock()

	refresh := s.IssueRefresh(tenant, senderDomains...)
	writeJSON(w, http.StatusOK, nms.Enrollment{
		Tenant:        tenant,
		Refresh:       refresh,
		SenderDomains: senderDomains,
		Limits:        nms.Limits{PerHour: 500, PerDay: 5000},
	})
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Refresh string `json:"refresh"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tokens++
	tenant, ok := s.refresh[req.Refresh]
	if !ok {
		writeError(w, http.StatusUnauthorized, nms.CodeRefreshInvalid, "unknown or revoked refresh token")
		return
	}

	delete(s.refresh, req.Refresh)
	refresh := "nmr_" + randHex()
	access := "nma_" + randHex()
	s.refresh[refresh] = tenant
	s.access[access] = tenant

	lifetime := s.AccessLifetime
	if lifetime <= 0 {
		lifetime = time.Hour
	}

	writeJSON(w, http.StatusOK, nms.Tokens{Access: access, ExpiresIn: int(lifetime.Seconds()), Refresh: refresh})
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	s.mutex.Lock()
	s.sends++
	tenant, ok := s.access[token]
	domains := s.tenants[tenant]
	s.mutex.Unlock()

	if !ok {
		writeError(w, http.StatusUnauthorized, nms.CodeTokenRevoked, "unknown or revoked access token")
		return
	}

	var req struct {
		Messages []nms.Message `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	var res []nms.MessageResult
	for _, msg := range req.Messages {
		if errs := validate(msg, domains); len(errs) > 0 {
			res = append(res, nms.MessageResult{Status: nms.StatusError, CustomID: msg.CustomID, Errors: errs})
			continue
		}

		s.mutex.Lock()
		s.sent = append(s.sent, msg)
		s.mutex.Unlock()
		res = append(res, nms.MessageResult{Status: nms.StatusQueued, CustomID: msg.CustomID, MessageID: "nms-" + randHex()})
	}

	writeJSON(w, http.StatusOK, map[string]any{"messages": res})
}

func validate(msg nms.Message, domains []string) []nms.MessageError {
	var errs []nms.MessageError
	_, domain, ok := strings.Cut(msg.From.Email, "@")
	if !ok {
		errs = append(errs, nms.MessageError{Code: nms.CodeInvalidSender, Field: "from.email"})
	} else if !contains(domains, domain) {
		errs = append(errs, nms.MessageError{Code: nms.CodeSenderNotAllowed, Field: "from.email"})
	}

	if len(msg.To)+len(msg.Cc)+len(msg.Bcc) == 0 {
		errs = append(errs, nms.MessageError{Code: nms.CodeInvalidRecipient, Field: "to"})
	}

	for i, to := range msg.To {
		if !strings.Contains(to.Email, "@") {
			errs = append(errs, nms.MessageError{Code: nms.CodeInvalidRecipient, Field: fmt.Sprintf("to[%d].email", i)})
		}
	}

	if msg.TextPart == "" && msg.HTMLPart == "" {
		errs = append(errs, nms.MessageError{Code: nms.CodeMissingBody})
	}

	return errs
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if strings.EqualFold(e, s) {
			return true
		}
	}

	return false
}

func randHex() string {
	var tmp [16]byte
	_, _ = rand.Read(tmp[:])
	return hex.EncodeToString(tmp[:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}
