package ora

import (
	"fmt"
	"strings"
)

// Color specifies the color, style or even semantics for the user when using a component.
// See also https://experience.sap.com/fiori-design-web/how-to-use-semantic-colors/.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Color string

const (
	// Primary call-to-action intention.
	Primary Color = "p"

	// Secondary call-to-action intention.
	Secondary Color = "s"

	// Tertiary call-to-action intention.
	Tertiary Color = "t"

	// Error describes a negative or a destructive intention. In Western Europe usually red. Use it, when the
	// user cannot continue normally and has to fix the problem first.
	Error Color = "n"

	// Warning describes a critical condition. In Western Europe usually yellow. Use it to warn on situations,
	// which may result in a future error condition.
	Warning Color = "c"

	// Positive describes a good condition or a confirming intention. In Western Europe usually green.
	// Use it to symbolize something which has been successfully applied.
	Positive Color = "o"

	// Informative shall be used to highlight something, which just changed. E.g. a newly added component or
	// a recommendation from a system. Do not use it to highlight text. In Western Europe usually blue.
	Informative Color = "i"

	// Regular shall be used for any default of any UI element which has no special semantic intention.
	// An empty color is always regular.
	Regular Color = "r"
)

// ExplicitColor accepts hex color codes (e.g. #1b8c30). Note, that these colors violates the WCAG accessibility
// guidelines and may even cause a legal dispute at worst (z.B. Abmahnung durch Wettbewerber).
// This function panics, if color looks invalid. See also [ParseNamedColor].
func ExplicitColor(color string) Color {
	c, err := ParseNamedColor(color)
	if err != nil {
		panic(err)
	}

	return c
}

// ParseNamedColor accepts currently anything starting with a # and treats it as a hex rgb(a) value.
// It also accepts all semantic ora color names.
func ParseNamedColor(color string) (Color, error) {
	if strings.HasPrefix(color, "#") {
		return Color(color), nil
	}

	switch Color(color) {
	case Primary:
		return Primary, nil
	case Secondary:
		return Secondary, nil
	case Tertiary:
		return Tertiary, nil
	case Error:
		return Error, nil
	case Warning:
		return Warning, nil
	case Positive:
		return Positive, nil
	case Informative:
		return Informative, nil
	case Regular:
		return Regular, nil
	case "":
		return Regular, nil

	default:
		return "", fmt.Errorf("unknown color '%s'", color)
	}
}
