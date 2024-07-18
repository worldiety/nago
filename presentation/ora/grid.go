package ora

// A Grid must support up to 12 Columns and a reasonable "unlimited" amount of rows.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Grid struct {
	Type    ComponentType `json:"type" value:"G"`
	Cells   []GridCell    `json:"b,omitempty"`
	Rows    int64         `json:"r,omitempty"`
	Columns int64         `json:"c,omitempty"`
	// InnerGap is omitted, if empty
	RowGap Length `json:"rg,omitempty"`
	ColGap Length `json:"cg,omitempty"`
	// Frame is omitted if empty
	Frame Frame `json:"f,omitempty"`
	// BackgroundColor regular is always transparent
	BackgroundColor Color   `json:"bgc,omitempty"`
	Padding         Padding `json:"p,omitempty"`
	Border          Border  `json:"bd,omitempty"`
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel string   `json:"al,omitempty"`
	Invisible          bool     `json:"iv,omitempty"`
	Font               Font     `json:"fn,omitempty"`
	ColWidths          []Length `json:"cw,omitempty"`
	component
}
