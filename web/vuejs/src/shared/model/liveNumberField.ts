import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyInt } from '@/shared/model/propertyInt';

export interface LiveNumberField {
	type: 'NumberField';
	id: number;
	label: PropertyString;
	hint: PropertyString;
	error: PropertyString;
	value: PropertyInt;
	placeholder: PropertyString;
	simple: PropertyBool;
	disabled: PropertyBool;
	onValueChanged: PropertyFunc;
}
