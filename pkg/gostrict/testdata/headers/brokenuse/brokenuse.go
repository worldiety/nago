package brokenuse

import "go.wdy.de/nago/pkg/gostrict/testdata/headers/broken"

func Run(x float64) float64 {
	return broken.Round(x)
}
