export interface LiveDialog {
	type: 'Dialog'
	id: number
	title: PropertyString
	body: PropertyComponent<LiveComponent>
	icon: PropertyString
	actions: ComponentList<LiveButton>
}
