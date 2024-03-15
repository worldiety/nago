export interface LivePage {
	type: 'Page'
	root: LiveComponent
	modals: ComponentList<LiveComponent>
	token: string
}
