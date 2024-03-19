import type { PropertyString } from '@/shared/model/propertyString';

export interface LiveStepInfo {
	type: 'StepInfo';
	id: number;
	number: PropertyString;
	caption: PropertyString;
	details: PropertyString;
}
