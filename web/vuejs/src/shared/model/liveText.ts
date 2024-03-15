export interface LiveText {
	type: 'Text'
	id: number
	value: PropertyString
	color: PropertyString
	colorDark: PropertyString
	size: PropertyString
	onClick: PropertyFunc
	onHoverStart: PropertyFunc
	onHoverEnd: PropertyFunc
}
