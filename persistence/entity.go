package persistence

import (
	"fmt"
	"go.wdy.de/nago/container/enum"
	"go.wdy.de/nago/container/serrors"
	"strconv"
)

// An Entity has an artificial identity.
type Entity[Ident comparable] interface {
	Identity() Ident
}

// EntityNotFound declares an error which describes that an existing entity identified by its Ident was expected
// but has not been found.
type EntityNotFound string // todo make this generic, when type aliases can be generic or RHS can be generic

func (e EntityNotFound) Error() string {
	return fmt.Sprintf("expected entity '%s' but it was not found", string(e))
}

// LookupFailure is an enum of two error situations.
type LookupFailure = enum.E2[EntityNotFound, serrors.InfrastructureError]

// IdentString converts the given identifier into a string representation.
func IdentString[Ident comparable](id Ident) string {
	switch t := any(id).(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
