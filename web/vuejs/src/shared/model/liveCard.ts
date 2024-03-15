export interface LiveCard {
	type: 'Card'
	children: ComponentList<LiveComponent>
	action: PropertyFunc
}
