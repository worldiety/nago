import type { ComponentList } from '@/shared/model/componentList';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveDropdown {
	type: 'Dropdown',
	id: number,
	items: ComponentList<LiveDropdownItem>,
	selectedIndex: PropertyInt,
	expanded: PropertyBool,
	disabled: PropertyBool,
	onToggleExpanded: PropertyFunc,
}
