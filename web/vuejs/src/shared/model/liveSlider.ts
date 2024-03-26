import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyString } from '@/shared/model/propertyString';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

export interface LiveSlider {
	type: 'Slider',
	id: number,
	disabled: PropertyBool,
	label: PropertyString,
	hint: PropertyString,
	error: PropertyString,
	value: PropertyInt,
	min: PropertyInt,
	max: PropertyInt,
	stepsize: PropertyInt,
	initialized: PropertyBool,
	onChanged: PropertyFunc,
}
