import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveTextField {
	type: 'TextField';
	id: number;
	label: PropertyString;
	hint: PropertyString;
	error: PropertyString;
	value: PropertyString;
	placeholder: PropertyString;
	disabled: PropertyBool;
	simple: PropertyBool;
	onTextChanged: PropertyFunc;
}
