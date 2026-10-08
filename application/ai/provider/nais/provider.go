// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nais

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/provider/anthropic"
	"go.wdy.de/nago/pkg/data/json"
)

const (
	// EnvService is the endpoint of the Nago AI Service. If not set, [DefaultEndpoint] is used. An empty value or
	// "off" disables the service.
	EnvService = "NAGO_AI_SERVICE"
	// EnvServiceToken is an optional refresh token of the Nago AI Service. Without it, the instance enrolls by a token
	// exchange: a public https origin is called back, any other origin like localhost must call from an address the
	// service knows.
	EnvServiceToken = "NAGO_AI_SERVICE_TOKEN"
)

// ID is the identity of the provider. It is fixed, because there is one service per instance, and conversations
// refer to their provider by it.
const ID provider.ID = "nago.ai.service"

var _ = registerService()

func registerService() any {
	provider.RegisterService("nais", newProvider)
	return nil
}

// endpointOf returns the configured endpoint or the default one, an empty string disables the service.
func endpointOf(env provider.ServiceEnv) string {
	v, ok := env.LookupEnv(EnvService)
	if !ok {
		return DefaultEndpoint
	}

	v = strings.TrimSpace(v)
	if strings.EqualFold(v, "off") {
		return ""
	}

	return v
}

// newProvider is the [provider.ServiceFactory] of the service. The provider is the Anthropic provider, pointed at the
// service and authenticated by the [Service].
func newProvider(env provider.ServiceEnv) (provider.Provider, bool, error) {
	endpoint := endpointOf(env)
	if endpoint == "" {
		return nil, false, nil
	}

	store, err := env.Store("nago.ai.nais")
	if err != nil {
		return nil, false, fmt.Errorf("cannot open the store of the nago ai service: %w", err)
	}

	nonces := NewNonces()
	svc := NewService(Options{
		Endpoint: endpoint,
		Token:    strings.TrimSpace(env.Getenv(EnvServiceToken)),
		Origin:   env.Origin,
		Nonces:   nonces,
		States:   json.NewSloppyJSONRepository[State, string](store),
	})

	env.HandleMethod(http.MethodGet, NoncePath+"{nonce}", nonces.Handler())
	slog.Info("nago ai service enabled", "endpoint", svc.Endpoint())

	return NewProvider(svc, nil), true, nil
}

// NewProvider returns the Anthropic provider speaking to the service, authenticated by svc. base is the transport
// underneath, nil means the default one.
func NewProvider(svc *Service, base http.RoundTripper) provider.Provider {
	if base == nil {
		base = http.DefaultTransport
	}

	return anthropic.NewProviderAt(ID, anthropic.Settings{
		Name:        "Nago AI Service",
		Description: "Models of the Nago AI Service at " + svc.Endpoint() + ". The instance enrolls by itself and needs no credentials.",
	}, anthropic.Endpoint{
		BaseURL: svc.Endpoint() + "/v1/",
		// large model responses can take a while, like with Anthropic itself
		HTTP: &http.Client{Timeout: 10 * time.Minute, Transport: transport{service: svc, base: base}},
	})
}
