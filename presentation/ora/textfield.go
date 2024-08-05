package ora

import "time"

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TextFieldStyle string

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TextField struct {
	Type           ComponentType `json:"type" value:"F"`
	Label          string        `json:"l,omitempty"`
	SupportingText string        `json:"s,omitempty"`
	// ErrorText is shown instead of SupportingText, even if they are (today) independent
	ErrorText string `json:"e,omitempty"`

	// Value contains the text, which shall be shown.
	Value string `json:"v,omitempty"`

	// InputValue is a binding to a state, into which the frontend will the user entered text. This is the pointer
	// to a [Property].
	InputValue Ptr  `json:"p"`
	Disabled   bool `json:"d,omitempty"`

	// Leading shows the given component usually at the left (or right if RTL). This can be used for additional
	// symbols like a magnifying glass for searching.
	Leading Component `json:"a,omitempty"`

	// Trailing show the given component usually at the right (or left if RTL mode). If set, the clear (or x button)
	// must not be shown, to reduce distraction. This can be used for an Info button or a text showing a value unit.
	Trailing Component `json:"r,omitempty"`

	// Style to apply. Use TextFieldReduced in forms where many textfields cause too much visual noise and you
	// need to reduce it. By default, the TextFieldOutlined is applied.
	Style TextFieldStyle `json:"t,omitempty"`

	// DebounceTime is in nanoseconds. A zero or omitted value means to enable debounce default logic.
	DebounceTime time.Duration `json:"dt,omitempty"`

	// Lines enforces a single line if <= 0, otherwise it shows the amount of text lines within a text area.
	Lines int `json:"li,omitempty"`

	// DisableDebounce must be set to true, to disable the default debouncer logic. This will cause a render roundtrip
	// for each keystroke, so be careful not to break the server or cause UX issues due to UI latencies.
	DisableDebounce bool  `json:"i,omitempty"`
	Invisible       bool  `json:"iv,omitempty"`
	Frame           Frame `json:"f,omitempty"`
	component
}
