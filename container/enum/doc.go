// deprecated: use github.com/worldiety/macro
// Package enum provides helper types to model type safe, serializable and exhaustive enums in Go.
// It is not as elegant as native type switches and they will be removed as soon as reasonable choice types
// are available in Go.
// This implementation uses a tagged union approach using an internal ordinal which relates to each specified
// generic type parameter. Since Go 1.21 we have a reasonable type inference, so that this approach becomes
// useable in practice.
package enum
