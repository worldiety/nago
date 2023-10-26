package application

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/container/slice"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type authProviders struct {
	keycloak *oidcProvider
}

// oidcProvider does not perform any oauth workflow in the backend. Instead, it expects that workflow at the
// frontend-side and only gets the jwts
type oidcProvider struct {
	Authority    string // e.g. http://localhost:8080/realms/myapp for a local keycloak or https://accounts.google.com
	ClientID     string // used by the frontend
	ClientSecret string // used by the frontend
	RedirectURL  string // used by the frontend
	Provider     *oidc.Provider
}

// KeycloakAuthentication enables the oauth support configuration using a keycloak instance.
// Your environment must provide AUTH_KEYCLOAK_REALM, AUTH_OIDC_CLIENT_ID, AUTH_OIDC_CLIENT_SECRET
// and AUTH_OIDC_REDIRECT_URL variables.
func (c *Configurator) KeycloakAuthentication() *Configurator {
	const (
		envRealm    = "AUTH_KEYCLOAK_REALM"
		envClid     = "AUTH_OIDC_CLIENT_ID"
		envSecret   = "AUTH_OIDC_CLIENT_SECRET"
		envRedirect = "AUTH_OIDC_REDIRECT_URL"
	)

	for _, envVar := range []string{envRealm, envClid, envSecret, envRedirect} {
		if os.Getenv(envVar) == "" {
			c.defaultLogger().Error("keycloak authentication required but required variable is missing", slog.String("env", envVar))
		}
	}

	providerInfo := &oidcProvider{
		Authority:    os.Getenv(envRealm),
		ClientID:     os.Getenv(envClid),
		ClientSecret: os.Getenv(envSecret),
		RedirectURL:  os.Getenv(""),
	}

	provider, err := oidc.NewProvider(c.Context(), providerInfo.Authority)
	if err != nil {
		panic(fmt.Errorf("cannot create oidc keycloak provider for authority '%s': %w", providerInfo.Authority, err))
	}

	providerInfo.Provider = provider

	c.auth.keycloak = providerInfo

	return c
}

func (c *Configurator) keycloakMiddleware(h http.Handler) http.Handler {
	if c.auth.keycloak == nil {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if c.validateToken(c.auth.keycloak, c.Context(), token) {
			//userInfo, err := c.auth.keycloak.Provider.UserInfo(c.Context(), oauth2.StaticTokenSource(token))
			// TODO ??? seems to work differently than our szhb stuff? should we just parse the token directly?
			var user KeycloakUser
			ctx := auth.WithContext(c.Context(), user)
			r = r.WithContext(ctx)
		}

		h.ServeHTTP(w, r)
	})
}

func (c *Configurator) validateToken(p *oidcProvider, ctx context.Context, token string) bool {
	verifier := p.Provider.Verifier(&oidc.Config{
		ClientID: p.ClientID,
	})

	_, err := verifier.Verify(ctx, token)
	if err != nil {
		c.defaultLogger().Error("cannot verify oidc token", slog.Any("err", err))
		return false
	}

	return true
}

type KeycloakUser struct {
	uid      string
	sid      string
	verified bool
	roles    []string
	mail     string
	name     string
}

func (k KeycloakUser) UID() string {
	return k.uid
}

func (k KeycloakUser) SID() string {
	return k.sid
}

func (k KeycloakUser) Verified() bool {
	return k.verified
}

func (k KeycloakUser) Roles() slice.Slice[string] {
	return slice.Of(k.roles...)
}

func (k KeycloakUser) ContactEmail() string {
	return k.mail
}

func (k KeycloakUser) ContactName() string {
	return k.name
}
