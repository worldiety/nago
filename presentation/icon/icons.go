package icon

import (
	_ "embed"
	"go.wdy.de/nago/presentation/ui"
)

//go:embed arrow-down.svg
var ArrowDown ui.SVGSrc

//go:embed arrow-up.svg
var ArrowUp ui.SVGSrc

//go:embed arrows-up-down.svg
var ArrowUpDown ui.SVGSrc
