import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveToggle {
	type: 'Toggle';
	id: number;
	label: PropertyString;
	checked: PropertyBool;
	disabled: PropertyBool;
	onCheckedChanged: PropertyFunc;
}
