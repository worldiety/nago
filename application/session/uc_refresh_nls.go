// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xhttp"
)

type nlsJSONAuthenticationDetails struct {
	Exp  int64   `json:"exp"`
	User nlsUser `json:"user"`
}

type nlsUser struct {
	ID                string   `json:"id"`
	BusinessPhones    []string `json:"businessPhones"`    // Office phone numbers
	DisplayName       string   `json:"displayName"`       // Full name
	GivenName         string   `json:"givenName"`         // First name
	Surname           string   `json:"surname"`           // Last name
	UserPrincipalName string   `json:"userPrincipalName"` // UPN, usually email
	Mail              string   `json:"mail"`              // Primary email
	JobTitle          string   `json:"jobTitle"`          // Job title
	MobilePhone       string   `json:"mobilePhone"`       // Mobile number
	OfficeLocation    string   `json:"officeLocation"`    // Office location
	PreferredLanguage string   `json:"preferredLanguage"` // e.g. "en-US"

	// Extended attributes (may not always be populated)
	City                        string   `json:"city"`
	Country                     string   `json:"country"`
	Department                  string   `json:"department"`
	EmployeeID                  string   `json:"employeeId"`
	FaxNumber                   string   `json:"faxNumber"`
	ImAddresses                 []string `json:"imAddresses"`
	MailNickname                string   `json:"mailNickname"`
	PostalCode                  string   `json:"postalCode"`
	State                       string   `json:"state"`
	StreetAddress               string   `json:"streetAddress"`
	UsageLocation               string   `json:"usageLocation"`
	AboutMe                     string   `json:"aboutMe"`
	AgeGroup                    string   `json:"ageGroup"`
	ConsentProvidedForMinor     string   `json:"consentProvidedForMinor"`
	LegalAgeGroupClassification string   `json:"legalAgeGroupClassification"`

	// Organization / company information
	CompanyName  string `json:"companyName"`
	EmployeeType string `json:"employeeType"`
}

func (u nlsUser) intoSSOUser() user.SingleSignOnUser {
	return user.SingleSignOnUser{
		// note: this is the stable object id of the identity provider (currently Entra), which the NLS just
		// passes through. It stays the same even if the mail address is changed.
		ID:                user.NLSUserID(u.ID),
		Firstname:         u.GivenName,
		Lastname:          u.Surname,
		Name:              u.DisplayName,
		Email:             user.Email(strings.ToLower(u.Mail)),
		PreferredLanguage: u.PreferredLanguage,
		Salutation:        "",
		Title:             u.JobTitle,
		Position:          u.JobTitle,
		CompanyName:       u.CompanyName,
		City:              u.City,
		PostalCode:        u.PostalCode,
		State:             u.State,
		Country:           u.Country,
		ProfessionalGroup: "",
		MobilePhone:       u.MobilePhone,
		AboutMe:           u.AboutMe,
	}
}

// ErrNLSUnavailable tells that the login service could not be asked or failed itself, e.g. due to a network
// problem or an outage of the identity provider. Unlike a rejected token, this keeps the session for a while.
var ErrNLSUnavailable = errors.New("login service unavailable")

const (
	// nlsGracePeriod is the time after the last successful refresh, for which a session survives an unavailable
	// login service. Afterward, it is logged out anyway, because a blocked account must not stay logged in.
	nlsGracePeriod = time.Hour

	// nlsAvatarInterval is the time after which the avatar is loaded again.
	nlsAvatarInterval = 24 * time.Hour

	nlsRefreshTimeout = 10 * time.Second
	nlsAvatarTimeout  = 5 * time.Second
)

// NewRefreshNLS asks the login service for the current user of the session and logs the session out, if the
// login service rejects its token. If the login service is unavailable, the session is kept for [nlsGracePeriod].
// Only the update of the session runs under the mutex, never a request.
func NewRefreshNLS(mutex *sync.Mutex, bus events.EventBus, repo Repository, loadGlobal settings.LoadGlobal, mergeUser user.MergeSingleSignOnUser, logout Logout) RefreshNLS {

	refresh := func(id ID) error {
		usrSettings := settings.ReadGlobal[user.Settings](loadGlobal)

		optSession, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("failed finding session: %w", err)
		}

		if optSession.IsNone() {
			return fmt.Errorf("session is gone: %s", logging.Secret(string(id)))
		}

		session := optSession.Unwrap()
		token := session.RefreshToken
		if token == "" {
			return fmt.Errorf("session has no nls refresh token: %s", logging.Secret(string(id)))
		}

		result, err := requestNLSRefresh(usrSettings.SSONLSServer, token)
		if err != nil {
			return err
		}

		if !user.Email(result.User.Mail).Valid() {
			return fmt.Errorf("invalid NLS email: '%s'", result.User.Mail)
		}

		now := nowFunc()
		var avatarBuf []byte
		avatarDue := now.Sub(session.NLSAvatarAt) >= nlsAvatarInterval
		if avatarDue {
			avatarBuf = loadAvatar(usrSettings.SSONLSServer, token)
		}

		uid, err := mergeUser(result.User.intoSSOUser(), avatarBuf)
		if err != nil {
			return fmt.Errorf("failed merging user: %w", err)
		}

		mutex.Lock()
		defer mutex.Unlock()

		optSession, err = repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("failed refreshing session: %w", err)
		}

		session = optSession.UnwrapOr(Session{})
		if session.RefreshToken != token {
			// logged out or signed in again in the meantime, which this refresh must not undo
			return nil
		}

		// Only the first sign-in sets the time of the authentication. A refresh every few minutes must not move it,
		// otherwise a single sign-on session would never reach its lifetime, see expired.
		if session.User.IsNone() || session.AuthenticatedAt.IsZero() {
			session.AuthenticatedAt = now
		}
		session.User = option.Some(uid)
		session.NLSRefreshedAt = now
		if avatarDue {
			session.NLSAvatarAt = now
		}

		if err := repo.Save(session); err != nil {
			return fmt.Errorf("failed saving session: %w", err)
		}

		slog.Info("nls refresh successful", "session", logging.Secret(string(id)), "user", uid)

		bus.Publish(Authenticated{
			Session: id,
			User:    session.User.Unwrap(),
		})

		return nil
	}

	// withinGracePeriod reports whether the session survives an unavailable login service. A session of an older
	// version has never been refreshed successfully, as far as it knows, and is logged out like before.
	withinGracePeriod := func(id ID) bool {
		optSession, err := repo.FindByID(id)
		if err != nil || optSession.IsNone() {
			return false
		}

		refreshedAt := optSession.Unwrap().NLSRefreshedAt
		return !refreshedAt.IsZero() && nowFunc().Sub(refreshedAt) < nlsGracePeriod
	}

	return func(id ID) error {
		err := refresh(id)
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrNLSUnavailable) && withinGracePeriod(id) {
			return err
		}

		if _, err2 := logout(id); err2 != nil {
			return fmt.Errorf("logout failed under failed condition: %w: %v", err2, err)
		}

		// not wrapped, the session is logged out and needs no retry
		return fmt.Errorf("failed refreshing session, logged out: %v", err)
	}
}

// requestNLSRefresh asks the login service for the current user. A rejection is a 4xx status. Everything else, like
// a network problem or a 5xx status, is [ErrNLSUnavailable].
func requestNLSRefresh(server string, token NLSRefreshToken) (nlsJSONAuthenticationDetails, error) {
	type body struct {
		Refresh NLSRefreshToken `json:"refresh"`
	}

	url := strings.TrimSuffix(server, "/") + "/api/nago/v1/refresh"
	buf := option.Must(json.Marshal(body{Refresh: token}))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nlsJSONAuthenticationDetails{}, fmt.Errorf("invalid nls refresh request: %s: %w", url, err)
	}

	client := &http.Client{Timeout: nlsRefreshTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nlsJSONAuthenticationDetails{}, fmt.Errorf("failed sending refresh request: %s: %w: %w", url, ErrNLSUnavailable, err)
	}

	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		return nlsJSONAuthenticationDetails{}, fmt.Errorf("nls rejected the refresh: %s: status %d", url, resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return nlsJSONAuthenticationDetails{}, fmt.Errorf("nls failed to refresh: %s: status %d: %w", url, resp.StatusCode, ErrNLSUnavailable)
	}

	var result nlsJSONAuthenticationDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nlsJSONAuthenticationDetails{}, fmt.Errorf("failed decoding refresh response: %s: %w: %w", url, ErrNLSUnavailable, err)
	}

	return result, nil
}

func loadAvatar(baseUrl string, token NLSRefreshToken) []byte {
	// grab optional image
	var buf []byte
	err := xhttp.NewRequest().
		BaseURL(baseUrl).
		URL("/api/nago/v1/photo/me").
		BearerAuthentication(string(token)).
		Timeout(nlsAvatarTimeout).
		Assert2xx(true).
		To(func(r io.Reader) error {
			tmp, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			buf = tmp
			return nil
		}).
		Get()

	if err != nil {
		slog.Warn("failed loading nls avatar", "err", err.Error())
		return nil
	}

	return buf
}
