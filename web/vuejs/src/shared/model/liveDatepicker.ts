import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveDatepicker {
	type: 'Dropdown',
	id: number,
	disabled: PropertyBool,
	label: PropertyString,
	hint: PropertyString,
	error: PropertyString,
	expanded: PropertyBool,
	onToggleExpanded: PropertyFunc,
}
