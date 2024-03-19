import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveTextArea {
	type: 'TextArea';
	id: number;
	label: PropertyString;
	hint: PropertyString;
	error: PropertyString;
	value: PropertyString;
	rows: PropertyInt;
	disabled: PropertyBool;
	onTextChanged: PropertyFunc;
}
