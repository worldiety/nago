export interface LiveDropdownItem {
	type: 'DropdownItem',
	itemIndex: PropertyInt,
	content: PropertyString,
	onSelected: PropertyFunc,
}
