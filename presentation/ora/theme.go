package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type NamespaceName string

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
