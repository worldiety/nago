// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nais

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

// NoncePath is the path below the origin, under which the service fetches the nonce of an exchange.
const NoncePath = "/api/nago/v1/ai/nonce/"

// NonceLifetime is the time the service has to call back.
const NonceLifetime = 5 * time.Minute

// Nonces holds the nonces of running exchanges. Each nonce is answered at most once.
type Nonces struct {
	mutex  sync.Mutex
	issued map[string]time.Time
	now    func() time.Time
}

func NewNonces() *Nonces {
	return &Nonces{issued: map[string]time.Time{}, now: time.Now}
}

// Issue returns a new random nonce.
func (n *Nonces) Issue() string {
	var tmp [32]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		panic(err)
	}

	nonce := hex.EncodeToString(tmp[:])

	n.mutex.Lock()
	defer n.mutex.Unlock()

	now := n.now()
	for k, at := range n.issued {
		if now.Sub(at) > NonceLifetime {
			delete(n.issued, k)
		}
	}

	n.issued[nonce] = now
	return nonce
}

// Consume returns true and forgets the nonce, if it has been issued and is not yet expired.
func (n *Nonces) Consume(nonce string) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	at, ok := n.issued[nonce]
	if !ok {
		return false
	}

	delete(n.issued, nonce)
	return n.now().Sub(at) <= NonceLifetime
}

// Forget drops the nonce, e.g. after a failed exchange.
func (n *Nonces) Forget(nonce string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	delete(n.issued, nonce)
}

// Handler answers the call back of the service under [NoncePath].
func (n *Nonces) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		idx := strings.Index(r.URL.Path, NoncePath)
		if idx < 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		nonce := r.URL.Path[idx+len(NoncePath):]
		if nonce == "" || strings.Contains(nonce, "/") || !n.Consume(nonce) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"nonce": nonce})
	}
}
