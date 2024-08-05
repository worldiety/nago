package ora

import "encoding/json"

// SVG contains the valid embeddable source of Scalable Vector Graphics.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SVG []byte

// MarshalJSON is for small hero svgs way more efficient than base64 encoding, e.g. 1,6kib vs 1,3kib
func (svg SVG) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(svg))
}
