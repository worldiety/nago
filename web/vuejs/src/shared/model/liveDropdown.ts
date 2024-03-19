import type { ComponentList } from '@/shared/model/componentList';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';
import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveDropdown {
	type: 'Dropdown',
	id: number,
	items: ComponentList<LiveDropdownItem>,
	selectedIndices: ComponentList<number>,
	multiselect: PropertyBool,
	expanded: PropertyBool,
	disabled: PropertyBool,
	label: PropertyString,
	hint: PropertyString,
	error: PropertyString,
	onClicked: PropertyFunc,
}
