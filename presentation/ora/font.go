package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Font struct {
	// Name of the font or family name as fallback. Extra fallback declarations are unspecified and must be comma
	// separated.
	Name string `json:"n,omitempty"`

	// Size of the font
	Size Length `json:"s,omitempty"`

	Style FontStyle `json:"t,omitempty"`

	Weight FontWeight `json:"w,omitempty"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type FontStyle string

const (
	ItalicFontStyle FontStyle = "i"
	NormalFontStyle           = "n"
)

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type FontWeight int

const (
	NormalFontWeight FontWeight = 400
	BoldFontWeight   FontWeight = 700
)
