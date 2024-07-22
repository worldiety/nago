package ora

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type NamespaceName string

var validColorNamespaceNameRegex = regexp.MustCompile(`ˆ[A-Za-z0-9_\-]+$`)

func (c NamespaceName) Valid() bool {
	return validColorNamespaceNameRegex.MatchString(string(c))
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Themes struct {
	Dark  Theme `json:"dark"`
	Light Theme `json:"light"`
	//HighContrast Theme `json:"highContrast"`
	//Protanopie   Theme `json:"protanopie"`
	//Deuteranopie Theme `json:"deuteranopie"`
	//Tritanopie   Theme `json:"tritanopie"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Theme struct {
	Colors  map[NamespaceName]map[string]Color `json:"colors"`
	Lengths Lengths                            `json:"lengths"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Lengths struct {
	CustomLengths map[string]Length `json:"customLengths"`
}

// CreateThemes takes the given 3 colors and generates all required themes from it.
// Remember the following color definitions:
// * primary is used for the overall color impression. Most areas colors are derived from different brightness values.
// * secondary is used for some accent colors and derived mostly by different transparency levels.
// * tertiary is the color for interactive elements like buttons.
func CreateThemes(primary, secondary, tertiary Color) Themes {
	themes := Themes{
		Dark: Theme{
			Colors: map[NamespaceName]map[string]Color{},
		},
		Light: Theme{
			Colors: map[NamespaceName]map[string]Color{},
		},
	}

	light := DefaultColors(Light, primary, secondary, tertiary)
	dark := DefaultColors(Dark, primary, secondary, tertiary)

	themes.Light.Colors[light.Namespace()] = ConvertColorSetToMap(light)
	themes.Dark.Colors[dark.Namespace()] = ConvertColorSetToMap(dark)

	return themes
}

func ConvertColorSetToMap(colorSet ColorSet) map[string]Color {
	// expensive but simple variant of going typesafe to arbitrary
	var res map[string]Color
	buf, err := json.Marshal(colorSet)
	if err != nil {
		panic(fmt.Errorf("unreachable: %w", err))
	}

	err = json.Unmarshal(buf, &res)
	if err != nil {
		panic(fmt.Errorf("unreachable: %w", err))
	}

	return res
}
