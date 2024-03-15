import type { LiveComponent } from '@/shared/model/liveComponent';

export interface ComponentList<T extends LiveComponent> {
	type: 'componentList'
	id: number
	value: T[]
}
