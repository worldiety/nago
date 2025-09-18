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
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
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

func NewRefreshNLS(mutex *sync.Mutex, repo Repository, loadGlobal settings.LoadGlobal, mergeUser user.MergeSingleSignOnUser, logout Logout) RefreshNLS {

	refresh := func(id ID) error {
		usrSettings := settings.ReadGlobal[user.Settings](loadGlobal)

		optSession, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("failed finding session: %w", err)
		}

		if optSession.IsNone() {
			return fmt.Errorf("session is gone: %s", id)
		}

		session := optSession.Unwrap()

		if session.RefreshToken == "" {
			return fmt.Errorf("session has no nls refresh token: %s", id)
		}

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		type body struct {
			Refresh NLSRefreshToken `json:"refresh"`
		}

		url := strings.TrimSuffix(usrSettings.SSONLSServer, "/") + "/api/nago/v1/refresh"

		buf := option.Must(json.Marshal(body{Refresh: session.RefreshToken}))
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
		if err != nil {
			return fmt.Errorf("invalid nls refresh request: %s: %w", url, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed sending refresh request: %s: %w", url, err)
		}

		defer resp.Body.Close()

		var result nlsJSONAuthenticationDetails
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed decoding refresh response: %s: %w", url, err)
		}

		if !user.Email(result.User.Mail).Valid() {
			return fmt.Errorf("invalid NLS email: %s", result.User.Mail)
		}

		mutex.Lock()
		defer mutex.Unlock()

		optSession, err = repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("failed refreshing session: %w", err)
		}

		if optSession.IsNone() {
			return fmt.Errorf("session refresh is gone: %s", id)
		}

		session = optSession.Unwrap()
		uid, err := mergeUser(result.User.intoSSOUser())
		if err != nil {
			return fmt.Errorf("failed merging user: %w", err)
		}

		session.AuthenticatedAt = time.Now()
		session.User = option.Some(uid)

		if err := repo.Save(session); err != nil {
			return fmt.Errorf("failed saving session: %w", err)
		}

		slog.Info("nls refresh successful", "session", id, "user", uid)

		return nil
	}

	return func(id ID) error {
		if err := refresh(id); err != nil {
			if _, err2 := logout(id); err2 != nil {
				return fmt.Errorf("logout failed under failed condition: %w: %w", err2, err)
			}

			return fmt.Errorf("failed refreshing session: %w", err)
		}

		return nil
	}
}
