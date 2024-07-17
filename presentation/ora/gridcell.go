package ora

// GridCell is undefined, if explicit row start/col start etc. is set and span values.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type GridCell struct {
	Type      ComponentType `json:"type" value:"C"`
	Body      Component     `json:"b,omitempty"`
	Alignment Alignment     `json:"a,omitempty"` // default (== empty) must stretch the element
	ColStart  int64         `json:"cs,omitempty"`
	ColEnd    int64         `json:"ce,omitempty"`
	RowStart  int64         `json:"rs,omitempty"`
	RowEnd    int64         `json:"re,omitempty"`
	ColSpan   int64         `json:"cp,omitempty"`
	RowSpan   int64         `json:"rp,omitempty"`
	component
}
