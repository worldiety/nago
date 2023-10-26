// Package main contains a rudimentary server that has a private endpoint that can only be accessed with a valid JWT.
// This is only used to demonstrate how private info could be protected.
package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/private", private)

	addr := "0.0.0.0:3000"
	fmt.Println("Listening on", addr)
	http.ListenAndServe(addr, nil)
}

func private(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")

	// Stuff like this should be handled by the router library of your choice.
	if r.Method == "OPTIONS" {
		return
	}

	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	ok, err := validateToken(ctx, token)
	if err != nil {
		fmt.Println("Internal error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("this string can only be read from the server when you are logged in"))
}

func validateToken(ctx context.Context, token string) (bool, error) {
	provider, err := oidc.NewProvider(ctx, "http://localhost:8080/realms/nago")
	if err != nil {
		return false, fmt.Errorf("create provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
	})

	_, err = verifier.Verify(ctx, token)
	if err != nil {
		fmt.Println("Verify:", err)
		return false, nil
	}

	return true, nil
}
