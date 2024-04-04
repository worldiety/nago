import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyInt } from '@/shared/model/propertyInt';

export interface LiveDatepicker {
	type: 'Dropdown',
	id: number,
	disabled: PropertyBool,
	label: PropertyString,
	hint: PropertyString,
	error: PropertyString,
	expanded: PropertyBool,
	rangeMode: PropertyBool,
	selectedStartDay: PropertyInt,
	selectedStartMonth: PropertyInt,
	selectedStartYear: PropertyInt,
	selectedEndDay: PropertyInt,
	selectedEndMonth: PropertyInt,
	selectedEndYear: PropertyInt,
	onClicked: PropertyFunc,
	onSelectionChanged: PropertyFunc,
}
