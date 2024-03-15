import type { ComponentList } from '@/shared/model/componentList';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import { PropertyString } from '@/shared/model/propertyString';

export interface LiveDropdown {
	type: 'Dropdown',
	id: number,
	items: ComponentList<LiveDropdownItem>,
	selectedIndex: PropertyInt,
	expanded: PropertyBool,
	disabled: PropertyBool,
	label: PropertyString,
	hint: PropertyString,
	error: PropertyString,
	onToggleExpanded: PropertyFunc,
}
