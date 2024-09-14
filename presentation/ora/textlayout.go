package ora

// A TextLayout places its content like a native Text would layout its words, using the same rules for word wrap
// and alignments. This allows to style inline-components individually. SwiftUI can do this using + on
// Text and Images. Jetpack has the concept of annotated strings.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TextLayout struct {
	Type     ComponentType `json:"type" value:"ts"`
	Children []Component   `json:"c,omitempty"`
	Border   Border        `json:"b,omitempty"`
	// Frame is omitted if empty
	Frame Frame `json:"f,omitempty"`

	// BackgroundColor regular is always transparent
	BackgroundColor Color   `json:"bgc,omitempty"`
	Padding         Padding `json:"p,omitempty"`
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel string        `json:"al,omitempty"`
	Invisible          bool          `json:"iv,omitempty"`
	Font               Font          `json:"fn,omitempty"`
	Action             Ptr           `json:"t,omitempty"`
	TextAlignment      TextAlignment `json:"a,omitempty"`

	component
}
