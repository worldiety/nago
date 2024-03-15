export interface LiveButton {
	type: 'Button'
	id: number
	caption: PropertyString
	preIcon: PropertyString
	postIcon: PropertyString
	color: PropertyString
	action: PropertyFunc
	disabled: PropertyBool
}
