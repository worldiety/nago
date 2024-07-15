package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"regexp"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

const (
	apiSlug     = "/api/v1/"
	apiUiSlug   = apiSlug + "ui/"
	apiPageSlug = apiUiSlug + "page/"
	apiAppSlug  = apiUiSlug + "application"
)

type PageID string

func (p PageID) Validate() error {
	if len(validPageIdRegex.FindAllStringSubmatch(string(p), -1)) != 1 {
		return fmt.Errorf("the id '%s' is invalid and must match the [a-z0-9_\\-{/}]+", string(p))
	}

	return nil
}

// deprecated use core.Application
type Application struct {
	Name        string
	Version     string
	Description string
	OIDC        []OIDCProvider //deprecated must be unified/abstracted away
	Components  map[ora.ComponentFactoryId]func(realm core.Window) core.View
}

// OIDCProvider does not perform any oauth workflow in the backend. Instead, it expects that workflow at the
// frontend-side and only gets the jwts
type OIDCProvider struct {
	Name                  string `json:"name"`
	Authority             string `json:"authority"`             // e.g. http://localhost:8080/realms/myapp for a local keycloak or https://accounts.google.com
	ClientID              string `json:"clientID"`              // used by the frontend
	ClientSecret          string `json:"clientSecret"`          // used by the frontend
	RedirectURL           string `json:"redirectURL"`           // used by the frontend
	PostLogoutRedirectUri string `json:"postLogoutRedirectUri"` // used by frontend
}
