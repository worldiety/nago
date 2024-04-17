package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type authProviders struct {
	keycloak *oidcProvider
}

// oidcProvider does not perform any oauth workflow in the backend. Instead, it expects that workflow at the
// frontend-side and only gets the jwts
type oidcProvider struct {
	Authority             string // e.g. http://localhost:8080/realms/myapp for a local keycloak or https://accounts.google.com
	ClientID              string // used by the frontend
	ClientSecret          string // used by the frontend
	RedirectURL           string // used by the frontend
	PostLogoutRedirectUri string // used by frontend
	Provider              *oidc.Provider
}

const OIDC_KEYCLOAK = "Keycloak"

// KeycloakAuthentication enables the oauth support configuration using a keycloak instance.
// Your environment must provide AUTH_KEYCLOAK_REALM and AUTH_OIDC_CLIENT_ID. Secret per AUTH_OIDC_CLIENT_SECRET
// is optional.
func (c *Configurator) KeycloakAuthentication() *Configurator {
	const (
		envRealm  = "AUTH_KEYCLOAK_REALM" // e.g. something like http://localhost:8080/realms/master
		envClid   = "AUTH_OIDC_CLIENT_ID" // e.g. something like testclientid
		envSecret = "AUTH_OIDC_CLIENT_SECRET"
	)

	for _, envVar := range []string{envRealm, envClid, envSecret} {
		if os.Getenv(envVar) == "" {
			c.defaultLogger().Error("keycloak authentication required but required variable is missing", slog.String("env", envVar))
		}
	}

	providerInfo := &oidcProvider{
		Authority:             os.Getenv(envRealm),
		ClientID:              os.Getenv(envClid),
		ClientSecret:          os.Getenv(envSecret),
		RedirectURL:           fmt.Sprintf("%s://%s:%d/oauth", c.getScheme(), c.getHost(), c.getPort()),
		PostLogoutRedirectUri: fmt.Sprintf("%s://%s:%d", c.getScheme(), c.getHost(), c.getPort()),
	}

	provider, err := oidc.NewProvider(c.Context(), providerInfo.Authority)
	if err != nil {
		panic(fmt.Errorf("cannot create oidc keycloak provider for authority '%s': %w", providerInfo.Authority, err))
	}

	providerInfo.Provider = provider

	c.auth.keycloak = providerInfo

	c.uiApp.OIDC = append(c.uiApp.OIDC, ui.OIDCProvider{
		Name:                  OIDC_KEYCLOAK,
		Authority:             providerInfo.Authority,
		ClientID:              providerInfo.ClientID,
		ClientSecret:          providerInfo.ClientSecret,
		RedirectURL:           providerInfo.RedirectURL,
		PostLogoutRedirectUri: providerInfo.PostLogoutRedirectUri,
	})
	return c
}

func (c *Configurator) keycloakMiddleware(h http.Handler) http.Handler {
	if c.auth.keycloak == nil {
		return h
	}

	// this is still required for rest apis etc.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token != "" {
			// we are protecting the access when parsing the page params, see also ui/utils.go
			user, err := validateToken(c.auth.keycloak, r.Context(), token)
			if err != nil {
				logging.FromContext(r.Context()).Error("cannot validate token", slog.Any("err", err))
			} else {
				ctx := auth.WithContext(r.Context(), user)
				r = r.WithContext(ctx)
			}
		}

		h.ServeHTTP(w, r)
	})
}

func validateToken(p *oidcProvider, ctx context.Context, token string) (auth.User, error) {
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
		birth:     time.Now(),
	}, nil
}

type KeycloakUser struct {
	userId    string
	sessionId string
	verified  bool
	roles     []string
	mail      string
	name      string
	birth     time.Time
}

func (k KeycloakUser) UserID() auth.UserID {
	return auth.UserID(k.userId)
}

func (k KeycloakUser) Roles(yield func(auth.RoleID) bool) {
	for _, role := range k.roles {
		if !yield(auth.RoleID(role)) {
			return
		}
	}
}

func (k KeycloakUser) Valid() bool {
	return k.birth.Add(10 * time.Minute).Before(time.Now()) // this is not what the JWT defines, however it is another security feature
}

func (k KeycloakUser) SessionID() string {
	return k.sessionId
}

func (k KeycloakUser) Verified() bool {
	return k.verified
}

func (k KeycloakUser) Email() string {
	return k.mail
}

func (k KeycloakUser) Name() string {
	return k.name
}

type invalidUser struct {
}

func (i invalidUser) UserID() auth.UserID {
	return ""
}

func (i invalidUser) Roles(yield func(auth.RoleID) bool) {
}

func (i invalidUser) SessionID() string {
	return ""
}

func (i invalidUser) Verified() bool {
	return false
}

func (i invalidUser) Email() string {
	return ""
}

func (i invalidUser) Name() string {
	return ""
}

func (i invalidUser) Valid() bool {
	return false
}
