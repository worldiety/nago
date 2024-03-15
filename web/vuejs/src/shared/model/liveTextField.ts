export interface LiveTextField {
	type: 'TextField'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	value: PropertyString
	disabled: PropertyBool
	onTextChanged: PropertyFunc
}
