import type { Event } from '@/shared/protocol/gen/event';

export default class Future {

	private readonly resolve: (event: Event) => void;
	private readonly reject: (event: Event) => void;
	private readonly monotonicRequestId: number;

	constructor(monotonicRequestId: number, resolve: (event: Event) => void, reject: (event: Event) => void) {
		this.resolve = resolve;
		this.reject = reject;
		this.monotonicRequestId = monotonicRequestId;
	}

	resolveFuture(event: Event): void {
		if (event.type === "ErrorOccurred") {
			console.log(`future ${this.monotonicRequestId} is rejected`)
			this.reject(event)
			return
		}

		this.resolve(event);
	}

	getRequestId(): number {
		return this.monotonicRequestId;
	}
}
