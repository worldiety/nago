export interface LiveDropdown {
	type: 'Dropdown',
	id: number,
	items: ComponentList<LiveDropdownItem>,
	selectedIndex: PropertyInt,
	expanded: PropertyBool,
	disabled: PropertyBool,
	onToggleExpanded: PropertyFunc,
}
