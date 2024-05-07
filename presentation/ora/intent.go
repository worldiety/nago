package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Intent string

const (
	// Primary call-to-action intention.
	Primary Intent = "primary"
	// Secondary call-to-action intention.
	Secondary Intent = "secondary"
	// Tertiary call-to-action intention.
	Tertiary Intent = "tertiary"
	// Negative or destructive intention.
	Negative Intent = "negative"
	Notice   Intent = "notice"
	// Positive or confirming intention.
	Positive    Intent = "positive"
	Informative Intent = "informative"
	Success     Intent = "success"
)
