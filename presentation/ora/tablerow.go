package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TableRow struct {
	Ptr   Ptr                   `json:"id"`
	Type  ComponentType         `json:"type" value:"TableRow"`
	Cells Property[[]TableCell] `json:"cells"`
	component
}
