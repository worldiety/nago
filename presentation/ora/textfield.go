package ora

import "time"

type TextFieldStyle string

const (
	// TextFieldReduced has no outlines and thus less visual disruption in larger forms.
	TextFieldReduced TextFieldStyle = "r"

	// TextFieldOutlined is fine for smaller forms and helps to identify where to put text in the form.
	TextFieldOutlined TextFieldStyle = "o"
)

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

	// DisableDebounce must be set to true, to disable the default debouncer logic. This will cause a render roundtrip
	// for each keystroke, so be careful not to break the server or cause UX issues due to UI latencies.
	DisableDebounce bool  `json:"i,omitempty"`
	Invisible       bool  `json:"iv,omitempty"`
	Frame           Frame `json:"f,omitempty"`
	component
}
