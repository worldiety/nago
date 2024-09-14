package tutorial_31_doc

import "embed"

//go:embed domain/**/*.go cmd/**/*.go
var Src embed.FS
