package application

import (
	"context"
	"errors"
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

		user, err := c.validateToken(c.auth.keycloak, r.Context(), token)
		if err != nil {
			c.defaultLogger().Error("validate token", "error", err)
		} else {
			ctx := auth.WithContext(r.Context(), user)
			r = r.WithContext(ctx)
		}

		h.ServeHTTP(w, r)
	})
}

func (c *Configurator) validateToken(p *oidcProvider, ctx context.Context, token string) (auth.User, error) {
	if len(token) == 0 {
		return nil, errors.New("empty token")
	}

	verifier := p.Provider.Verifier(&oidc.Config{
		ClientID:          p.ClientID,
		SkipClientIDCheck: true, // TODO Figure out why this is required
	})

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	type Claims struct {
		SID           string `json:"sid"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Email         string `json:"email"`
	}
	var claims Claims

	err = idToken.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("get claims: %w", err)
	}

	return &KeycloakUser{
		userId:    idToken.Subject,
		sessionId: claims.SID,
		verified:  claims.EmailVerified,
		roles:     nil,
		mail:      claims.Email,
		name:      claims.Name,
	}, nil
}

type KeycloakUser struct {
	userId    string
	sessionId string
	verified  bool
	roles     []string
	mail      string
	name      string
}

func (k KeycloakUser) UserID() string {
	return k.userId
}

func (k KeycloakUser) SessionID() string {
	return k.sessionId
}

func (k KeycloakUser) Verified() bool {
	return k.verified
}

func (k KeycloakUser) Roles() slice.Slice[string] {
	return slice.Of(k.roles...)
}

func (k KeycloakUser) Email() string {
	return k.mail
}

func (k KeycloakUser) Name() string {
	return k.name
}
