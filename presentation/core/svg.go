package core

// SVG contains the valid embeddable source of Scalable Vector Graphics.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SVG []byte

func (svg SVG) AsBytes() []byte {
	return svg
}
