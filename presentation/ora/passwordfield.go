package ora

type PasswordField struct {
	Ptr               Ptr              `json:"id"`
	Type              ComponentType    `json:"type" value:"PasswordField"`
	Label             Property[string] `json:"label"`
	Hint              Property[string] `json:"hint"`
	Help              Property[string] `json:"help"`
	Error             Property[string] `json:"error"`
	Value             Property[string] `json:"value"`
	Revealed          Property[bool]   `json:"revealed"`
	Placeholder       Property[string] `json:"placeholder"` // TODO that does not make any sense from UX, we have Label and Hint: remove me
	Disabled          Property[bool]   `json:"disabled"`
	Simple            Property[bool]   `json:"simple"` // TODO what is that? Better use a documented enum?
	OnPasswordChanged Property[Ptr]    `json:"onPasswordChanged"`
	component
}
