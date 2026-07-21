package slug

import (
	"strings"

	"github.com/gosimple/slug"
)

func Slugify(input ...string) string {
	joined := strings.Join(input, "-")
	return slug.Make(joined)
}
