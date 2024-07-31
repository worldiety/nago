package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type DatePickerStyle string

const (
	DatePickerSingleDate DatePickerStyle = "s"
	DatePickerDateRange  DatePickerStyle = "r"
)

// Date represents a location-free representation of a day/month/year tuple.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Date struct {
	Day   int `json:"d,omitempty"`
	Month int `json:"m,omitempty"`
	Year  int `json:"y,omitempty"`
}

func (d Date) Zero() bool {
	return d == Date{}
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type DatePicker struct {
	Type           ComponentType `json:"type" value:"P"`
	Disabled       bool          `json:"d,omitempty"`
	Label          string        `json:"l,omitempty"`
	SupportingText string        `json:"s,omitempty"`
	// ErrorText is shown instead of SupportingText, even if they are (today) independent
	ErrorText string `json:"e,omitempty"`

	// Style determines if the picker shall use the range or single mode. Default is single selection
	Style DatePickerStyle `json:"y,omitempty"`

	// Value is the initial single value or start value of the picker.
	Value Date `json:"v"`

	// InputValue is the picked single value or end value of the picker.
	InputValue Ptr `json:"p"`

	// EndValue is the initial end value of the picker.
	EndValue Date `json:"ev"`

	// EndInputValue is the picked end value of the picker.
	EndInputValue Ptr `json:"ep"`

	Invisible bool `json:"iv,omitempty"`
	component
}
