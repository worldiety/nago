import type { Event } from '@/shared/protocol/gen/event';

export default class Future<T extends Event> {

	private readonly resolve: (response: T) => void;
	private readonly reject: (reason: unknown) => void;
	private readonly monotonicRequestId: number;

	constructor(monotonicRequestId: number, resolve: (response: T) => void, reject: (reason: unknown) => void) {
		this.resolve = resolve;
		this.reject = reject;
		this.monotonicRequestId = monotonicRequestId;
	}

	resolveFuture(response: T): void {
		if (response.type === "ErrorOccurred") {
			console.log(`future ${this.monotonicRequestId} is rejected`)
			this.reject(response)
			return
		}

		this.resolve(response);
	}

	getRequestId(): number {
		return this.monotonicRequestId;
	}
}
