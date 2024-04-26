import type { Event } from '@/shared/protocol/gen/event';
import { EventType } from '@/shared/eventbus/eventType';

type EventCallback = (event: Event) => void;

export default class EventBus {

	private readonly callbacksMap: Map<EventType, EventCallback[]> = new Map();

	subscribe(eventType: EventType, callback: EventCallback): void {
		const callbacks = this.callbacksMap.get(eventType) ?? [];
		callbacks.push(callback);
		this.callbacksMap.set(eventType, callbacks);
	}

	unsubscribe(eventType: EventType, callback: EventCallback): void {
		const callbacks = this.callbacksMap.get(eventType) ?? [];
		const existingCallbackIndex = callbacks.findIndex((existingCallback) => existingCallback === callback);
		if (existingCallbackIndex >= 0) {
			callbacks.splice(existingCallbackIndex, 1);
			this.callbacksMap.set(eventType, callbacks);
		}
	}

	publish(eventType: EventType, event: Event): void {
		const callbacks = this.callbacksMap.get(eventType) ?? [];
		callbacks.forEach(callback => callback(event));
	}
}
