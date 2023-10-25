package persistence

import (
	"fmt"
	"go.wdy.de/nago/container/enum"
	"strconv"
)

// An Entity has an artificial identity.
type Entity[Ident comparable] interface {
	Identity() Ident
}

// InfrastructureError is something like a broken I/O connection, disk full etc. which depends on the actual storage
// system. Even though these are all anticipated errors, the user cannot usually do something about it.
// Thus, the system mostly fails with an internal server error (500) at the presentation side.
// The responsibility to fix that is up to the service administrator.
type InfrastructureError interface {
	error
	isInfrastructure()
	Unwrap() error
}

type infrErr struct {
	Cause error
}

func (e infrErr) Unwrap() error {
	return e.Cause
}

func (e infrErr) isInfrastructure() {

}

func (e infrErr) Error() string {
	return "infrastructure error: " + e.Cause.Error()
}

// IntoInfrastructure returns an InfrastructureError if e is not nil. Otherwise, returns also nil.
func IntoInfrastructure(e error) InfrastructureError {
	if e == nil {
		return nil
	}

	return infrErr{Cause: e}
}

// EntityNotFound declares an error which describes that an existing entity identified by its Ident was expected
// but has not been found.
type EntityNotFound string // todo make this generic, when type aliases can be generic or RHS can be generic

func (e EntityNotFound) Error() string {
	return fmt.Sprintf("expected entity '%s' but it was not found", string(e))
}

// LookupFailure is an enum of two error situations.
type LookupFailure = enum.E2[EntityNotFound, InfrastructureError]

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
