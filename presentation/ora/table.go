package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Table struct {
	Ptr     Ptr                   `json:"id"`
	Type    ComponentType         `json:"type" value:"Table"`
	Headers Property[[]TableCell] `json:"headers"`
	Rows    Property[[]TableRow]  `json:"rows"`
	component
}
