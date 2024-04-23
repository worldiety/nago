export default class Future<T> {

	private readonly resolve: (responseRaw: T) => void;
	private readonly reject: (reason: unknown) => void;
	private readonly monotonicRequestId: number;

	constructor(monotonicRequestId: number, resolve: (response: T) => void, reject: (reason: unknown) => void) {
		this.resolve = resolve;
		this.reject = reject;
		this.monotonicRequestId = monotonicRequestId;
	}

	resolveFuture(responseRaw: T): void {
		if (responseRaw.type === "ErrorOccurred") {
			console.log(`future ${this.monotonicRequestId} is rejected`)
			this.reject(responseRaw)
			return
		}

		this.resolve(responseRaw);
	}

	getRequestId(): number {
		return this.monotonicRequestId;
	}
}
