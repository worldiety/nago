export interface LiveStepper {
	type: 'Stepper'
	id: number
	steps: ComponentList<LiveStepInfo>
	selectedIndex: PropertyInt
}
