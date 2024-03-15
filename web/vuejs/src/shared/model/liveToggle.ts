export interface LiveToggle {
	type: 'Toggle'
	id: number
	label: PropertyString
	checked: PropertyBool
	disabled: PropertyBool
	onCheckedChanged: PropertyFunc
}
