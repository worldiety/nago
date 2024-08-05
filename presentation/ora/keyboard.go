package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type KeyboardOptions struct {
	Capitalization     bool         `json:"c,omitempty"`
	AutoCorrectEnabled bool         `json:"a,omitempty"`
	KeyboardType       KeyboardType `json:"k,omitempty"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type KeyboardType string

const (
	KeyboardDefault KeyboardType = ""
	KeyboardAscii   KeyboardType = "a"
	KeyboardInteger KeyboardType = "i"
	KeyboardFloat   KeyboardType = "f"
	KeyboardEMail   KeyboardType = "m"
	KeyboardPhone   KeyboardType = "p"
	KeyboardSearch  KeyboardType = "s"
	KeyboardURL     KeyboardType = "u"
)
