export interface LiveScaffold {
	type: 'Scaffold'
	id: number
	title: PropertyString
	breadcrumbs: ComponentList<LiveButton> // currently ever of LiveButton
	menu: ComponentList<LiveButton> // currently always of LiveButton
	body: PropertyComponent<LiveComponent>
	topbarLeft: PropertyComponent<LiveComponent>
	topbarMid: PropertyComponent<LiveComponent>
	topbarRight: PropertyComponent<LiveComponent>
}
