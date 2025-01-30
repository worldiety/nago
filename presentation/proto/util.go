package proto

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var validComponentIdRegex = regexp.MustCompile(`[A-Za-z0-9_\-{/}]+`)

// NewScopeID returns a hex encoded and filename safe 32 byte entropy.
func NewScopeID() ScopeID {
	var tmp [32]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		panic(err)
	}

	return ScopeID(hex.EncodeToString(tmp[:]))
}

// Valid identifies a unique constructor for a specific ComponentType.
// Such an addressable Component is likely a page and instantiated and rendered.
// In return, a ComponentInvalidated event will be sent in the future.
// For details, see the [NewComponentRequested] event.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
func (c RootViewID) Valid() bool {
	if c == "." {
		return true
	}

	if strings.HasPrefix(string(c), "/") || strings.HasSuffix(string(c), "/") {
		return false
	}

	return validComponentIdRegex.FindString(string(c)) == string(c)
}
