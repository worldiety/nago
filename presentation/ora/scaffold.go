package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ScaffoldMenuEntry struct {
	Icon       Component           `json:"i,omitempty"`
	IconActive Component           `json:"v,omitempty"`
	Title      string              `json:"t,omitempty"`
	Action     Ptr                 `json:"a,omitempty"`
	Factory    ComponentFactoryId  `json:"f,omitempty"`
	Menu       []ScaffoldMenuEntry `json:"m,omitempty"`
	Badge      string              `json:"b,omitempty"`
	Expanded   bool                `json:"x,omitempty"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ScaffoldAlignment string

const (
	ScaffoldAlignmentTop     ScaffoldAlignment = "u"
	ScaffoldAlignmentLeading ScaffoldAlignment = "l"
)

// Scaffold is only defined as a root view. Other use cases are undefined and will likely break the rendering.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Scaffold struct {
	Type ComponentType       `json:"type" value:"A"`
	Body Component           `json:"b,omitempty"`
	Logo Component           `json:"l,omitempty"`
	Menu []ScaffoldMenuEntry `json:"m,omitempty"`
	// Alignment defaults to Leading (usually Left).
	Alignment ScaffoldAlignment `json:"a,omitempty"`

	component
}
