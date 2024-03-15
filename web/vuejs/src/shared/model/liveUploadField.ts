export interface LiveUploadField {
	type: 'FileField'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	disabled: PropertyBool
	filter: PropertyString
	multiple: PropertyBool
	uploadToken: PropertyString
}
