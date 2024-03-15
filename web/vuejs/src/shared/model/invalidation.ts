export interface Invalidation {
	type: 'Invalidation' | 'HistoryPushState' | 'HistoryBack'
	root: LiveComponent
	modals: ComponentList<LiveComponent>
	token: string
}
