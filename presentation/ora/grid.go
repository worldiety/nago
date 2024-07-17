package ora

// A Grid must support up to 12 Columns and a reasonable "unlimited" amount of rows.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Grid struct {
	Type    ComponentType `json:"type" value:"G"`
	Cells   []GridCell    `json:"b,omitempty"`
	Rows    int64         `json:"r,omitempty"`
	Columns int64         `json:"c,omitempty"`
	Gap     Length        `json:"g,omitempty"`
	component
}
