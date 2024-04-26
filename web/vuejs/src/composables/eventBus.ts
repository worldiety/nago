import { inject } from 'vue';
import { eventBusKey } from '@/shared/injectionKeys';
import type EventBus from '@/shared/eventbus/eventBus';

export function useEventBus(): EventBus {
	const eventBus = inject(eventBusKey);
	if (!eventBus) {
		throw new Error('Could not inject EventBus as it is undefined');
	}

	return eventBus;
}
