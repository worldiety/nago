package ora

// Table represents a pre-styled table with limited styling capabilities. Use Grid for maximum flexibility.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Table struct {
	Type               ComponentType `json:"type" value:"B"`
	Header             TableHeader   `json:"h,omitempty"`
	Rows               []TableRow    `json:"r,omitempty"`
	Frame              Frame         `json:"f,omitempty"`
	BackgroundColor    Color         `json:"bgc,omitempty"`
	Border             Border        `json:"b,omitempty"`
	DefaultCellPadding Padding       `json:"p,omitempty"`
	RowDividerColor    Color         `json:"rdc,omitempty"`
	component
}

// TableHeader aggregates the optional header properties and defines columns from left to right.
// We are not assigning cells to columns by id, to lower the protocol overhead.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TableHeader struct {
	Columns []TableColumn `json:"c,omitempty"`
}

// TableRow p
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TableRow struct {
	Cells           []TableCell `json:"c,omitempty"`
	Height          Length      `json:"h,omitempty"`
	BackgroundColor Color       `json:"b,omitempty"`
	Action          Ptr         `json:"a,omitempty"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TableCell struct {
	Content Component `json:"c,omitempty"`
	// Values higher than 65534 are clipped.
	RowSpan int `json:"rs,omitempty"`
	// Values higher than 1000 are clipped.
	ColSpan         int       `json:"cs,omitempty"`
	Alignment       Alignment `json:"a,omitempty"`
	BackgroundColor Color     `json:"b,omitempty"`
	Padding         Padding   `json:"p,omitempty"`
	Border          Border    `json:"o,omitempty"`
	Action          Ptr       `json:"t,omitempty"`
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TableColumn struct {
	Content Component `json:"c,omitempty"`
	// Values higher than 1000 are clipped.
	ColSpan         int       `json:"cs,omitempty"`
	Width           Length    `json:"w,omitempty"`
	Alignment       Alignment `json:"a,omitempty"`
	BackgroundColor Color     `json:"b,omitempty"`
	CellAction      Ptr       `json:"t,omitempty"`
}
