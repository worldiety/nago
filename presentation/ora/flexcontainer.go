package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type FlexContainer struct {
	Ptr              Ptr                        `json:"id"`
	Type             ComponentType              `json:"type" value:"FlexContainer"`
	Elements         Property[[]Component]      `json:"elements"`
	ElementSize      Property[ElementSize]      `json:"elementSize"`
	Orientation      Property[Orientation]      `json:"orientation"`
	ContentAlignment Property[ContentAlignment] `json:"contentAlignment"`
	ItemsAlignment   Property[ItemsAlignment]   `json:"itemsAlignment"`
	Visible          Property[bool]             `json:"visible"`
	MaxWidth         Property[ElementSize]      `json:"maxWidth"`
	component
}
