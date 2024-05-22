package data

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/std"
	"reflect"
	"strconv"
)

// SkipAll is a sentinel error for traversals.
var SkipAll = errors.New("skip everything and stop the traversal")

// MapF defines a generic mapping function
type MapF[From, To any] func(From) (To, error) // TODO this is not MapF but somewhat a MapF12

// An Aggregate is a special Entity but manages other entities or value types (aggregate root).
// It defines the boundary of a transaction, so if something must be consistent, it must be an aggregate root.
// A simple CRUD system is a special case, where the AggregateRoot is just an Entity.
type Aggregate[Ident comparable] interface {
	Identity() Ident
}

type IDType interface {
	~int | ~int64 | ~int32 | ~string
}

// Repository defines how to work with entities or more specific with aggregate roots, in the definition of
// domain driven design. A repository is a design pattern to separate the domain layer from the persistence layer
// implementation. In general, this pattern improves maintainability and testability of the overall code base.
// Other repository implementations can be swapped without touching the domain,
// as requirements regarding the persistence changes.
//
// Note, that most implementations may not tolerate nested repository calls.
// For example, calling functions from a yield on the same instance will likely cause a deadlock or other
// incorrect or unexpected results.
type Repository[E Aggregate[ID], ID IDType] interface {
	// FindByID either returns some Aggregate or none.
	// The effort is implementation dependent, but most reasonable implementations guarantee something better than
	// O(n) like a tree O(log(n) or even a hashtable (O(1)).
	// Returned errors are unspecified infrastructure errors of the implementation.
	FindByID(id ID) (std.Option[E], error)

	// FindAllByID collects all available entities and yields at most (or less) than the amount of given ids.
	// It is not an error, if an entity has not been found.
	// The order is undefined, to allow optimizations.
	// Implementations with transaction support must find within a single transaction.
	// Returned errors are unspecified infrastructure errors of the implementation.
	// The yield signature intentionally matches the according [iter.Seq2] part. See also [Repository.Each].
	FindAllByID(ids iter.Seq[ID], yield func(E, error) bool)

	// DeleteByID removes the specified aggregate. If no such entity exists, no error is returned.
	// Returned errors are unspecified infrastructure errors of the implementation.
	DeleteByID(id ID) error

	// DeleteAll removes all entities at some point in time. Just like count, this depends on concurrent usage
	// and pending transactions.
	// Implementations with transaction support must delete within a single transaction.
	// Returned errors are unspecified infrastructure errors of the implementation.
	DeleteAll() error

	// DeleteAllByID is like DeleteAll and DeleteByID but using the given sequence of identifiers.
	// Implementations with transaction support must delete within a single transaction.
	// Returned errors are unspecified infrastructure errors of the implementation.
	DeleteAllByID(ids iter.Seq[ID]) error

	// Delete removes entities by predicate, which is always O(n) due to a full traversal.
	// Implementations with transaction support must delete within a single transaction.
	// Returned errors are unspecified infrastructure errors of the implementation.
	// If the predicate returns [SkipAll], the implementation must stop the traversal and return without error.
	Delete(predicate func(E) (bool, error)) error

	// Each loops over each item without any particular order until the callee returns false or all entries have
	// been visited.
	// The order of traversal is undefined and may be even different between subsequent calls.
	// Implementations with transaction support must find within a single transaction.
	// Returned errors are unspecified infrastructure errors of the implementation.
	// This is a [iter.Seq2].
	Each(yield func(E, error) bool)

	// Count returns the estimated amount of entries. Due to concurrent usage or pending transaction, this
	// value is only a kind of snapshot view and a subsequent iteration of all values may return
	// a different result.
	// The effort is implementation dependent and may be anything from O(n) to O(1).
	// Returned errors are unspecified infrastructure errors of the implementation.
	Count() (int, error)

	// Save persists a single aggregate.
	// Implementations with transaction support must save within a single transaction, hopefully Aora.Ptr.
	// Returned errors are unspecified infrastructure errors of the implementation.
	Save(E) error

	// SaveAll persists all given aggregates at once.
	// Implementations with transaction support must save all aggregates within a single transaction.
	// Returned errors are unspecified infrastructure errors of the implementation.
	SaveAll(it iter.Seq[E]) error
}

// RandIdent create a cryptographic secure random string containing 16 bytes of entropy.
// It is hex encoded and looks like 5134b3c04a7bbc56ab1b9435acfd98cb.
// In the future, we may increase this to 24 or 32 byte of entropy.
//
// Intentionally,
// we do not allow the number types, because they contain just 4-8 byte of entropy, which
// causes a large collision probability.
func RandIdent[T ~string]() T {
	var tmp [16]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		panic(err)
	}

	return T(hex.EncodeToString(tmp[:]))
}

// Idtos converts any IDType into a string. See Stoid for its inverse.
func Idtos[ID IDType](id ID) string {
	v := reflect.ValueOf(id)
	switch v.Kind() {
	case reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Int:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Int32:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.String:
		return v.String()
	default:
		panic(fmt.Errorf("unsupported id type: %T", id))
	}
}

// Stoid tries to parse a string value into the actual ID type. See Idtos for the inverse.
func Stoid[ID IDType](value string) (ID, error) {
	var zero ID
	v := reflect.ValueOf(zero)
	switch v.Kind() {
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 64)
		reflect.ValueOf(&zero).Elem().SetInt(i)
		return zero, err
	case reflect.String:
		reflect.ValueOf(&zero).Elem().SetString(value)
		return zero, nil
	default:
		panic(fmt.Errorf("unsupported id type: %T", zero))
	}
}
