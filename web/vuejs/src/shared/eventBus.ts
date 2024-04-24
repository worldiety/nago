import type { Event } from '@/shared/protocol/gen/event';

export default class EventBus {

	private readonly callbacksMap: Map<EventType, ((event: Event) => void)[]>;

	constructor() {
		this.callbacksMap = new Map();
	}

	subscribe(eventType: EventType, callback: (event: Event) => void): void {
		const callbacks = this.callbacksMap.get(eventType) ?? [];
		callbacks.push(callback);
		this.callbacksMap.set(eventType, callbacks);
	}

	unsubscribe(eventType: EventType, callback: (event: Event) => void): void {
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

export enum EventType {
	INVALIDATION,
}
