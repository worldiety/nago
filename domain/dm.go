package dm

import (
	"crypto/rand"
	"encoding/hex"
)

func NewID() string {
	var tmp [16]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		panic(err)
	}

	return hex.EncodeToString(tmp[:])
}
