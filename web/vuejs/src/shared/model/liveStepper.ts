import type { ComponentList } from '@/shared/model/componentList';
import type { LiveStepInfo } from '@/shared/model/liveStepInfo';
import type { PropertyInt } from '@/shared/model/propertyInt';

export interface LiveStepper {
	type: 'Stepper'
	id: number
	steps: ComponentList<LiveStepInfo>
	selectedIndex: PropertyInt
}
