package enum

// see also https://serde.rs/enum-representations.html#adjacently-tagged.
// we don't use the default variant, because it is ineffcient to express that naturally in go due
// to map usage.
type adjacentlyTagged[T any] struct {
	Type  string `json:"type,omitempty"`
	Value T      `json:"value,omitempty"`
}

type adjacentlyTaggedPreflight struct {
	Type string `json:"type"`
}
