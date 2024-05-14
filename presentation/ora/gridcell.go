package ora

// TODO this entire type is so HTML like and hard to handle and port to mobile devices. It has no semantics.
//
//	I vote for deletion, but what is the replacement?
//
// deprecated
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type GridCell struct {
	Ptr       Ptr                 `json:"id"`
	Type      ComponentType       `json:"type" value:"GridCell"`
	Body      Property[Component] `json:"body"`
	ColStart  Property[int64]     `json:"colStart"`
	ColEnd    Property[int64]     `json:"colEnd"`
	RowStart  Property[int64]     `json:"rowStart"`
	RowEnd    Property[int64]     `json:"rowEnd"`
	ColSpan   Property[int64]     `json:"colSpan"`
	SmColSpan Property[int64]     `json:"smColSpan"`
	MdColSpan Property[int64]     `json:"mdColSpan"`
	LgColSpan Property[int64]     `json:"lgColSpan"`
	component
}
