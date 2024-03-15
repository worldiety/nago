export interface LiveTextArea {
	type: 'TextArea'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	value: PropertyString
	rows: PropertyInt
	disabled: PropertyBool
	onTextChanged: PropertyFunc
}
