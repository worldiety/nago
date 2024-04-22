package ora

type Image struct {
	Ptr           Ptr              `json:"id"`
	Type          ComponentType    `json:"type" value:"Image"`
	URL           Property[string] `json:"url"`
	DownloadToken Property[string] `json:"downloadToken"`
	Caption       Property[string] `json:"caption"`
	component
}
