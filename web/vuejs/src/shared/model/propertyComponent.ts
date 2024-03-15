export interface PropertyComponent<T extends LiveComponent> {
	id: number
	name: string
	value: T
}
