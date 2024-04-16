package ora

// TODO this entire type is so HTML like and hard to handle and port to mobile devices. It has no semantics.
//
//	I vote for deletion, but what is the replacement?
//
// deprecated
type Grid struct {
	Ptr       Ptr                  `json:"id"`
	Type      ComponentType        `json:"type" value:"Grid"`
	Cells     Property[[]GridCell] `json:"cells"`
	Rows      Property[int64]      `json:"rows"`
	Columns   Property[int64]      `json:"columns"`
	SMColumns Property[int64]      `json:"smColumns"`
	MDColumns Property[int64]      `json:"mdColumns"`
	LGColumns Property[int64]      `json:"lgColumns"`
	Gap       Property[string]     `json:"gap"`
	component
}
