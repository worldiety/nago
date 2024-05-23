package ora

// SVG contains the valid embeddable source of Scalable Vector Graphics.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SVG string

func (svg SVG) AsBytes() []byte {
	return []byte(svg)
}

// RIDSVG is a Resource IDentifier for a Scalable Vector Graphics.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type RIDSVG int64
