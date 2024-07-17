package ora

// Spacer grows or shrinks within a HStack or VStack. In other layouts, the behavior is unspecified.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Spacer struct {
	Type            ComponentType `json:"type" value:"s"`
	Frame           Frame         `json:"f"`
	Border          Border        `json:"b"`
	BackgroundColor Color         `json:"bgc"`
	component
}
