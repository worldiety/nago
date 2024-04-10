import type { Invalidation } from '@/shared/model/invalidation';
import type NetworkAdapter from '@/shared/network/networkAdapter';
import type { Property } from '@/shared/model/property';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import { v4 as uuidv4 } from 'uuid';

export default class NetworkProtocol {

	private networkAdapter: NetworkAdapter;
	private pendingRequests: Map<string, Future>;

	constructor(networkAdapter: NetworkAdapter) {
		this.networkAdapter = networkAdapter;
		this.pendingRequests = new Map<string, Future>();
		this.networkAdapter.subscribe((responseRaw) => {
			const responseParsed = JSON.parse(responseRaw);
			const requestId = responseParsed['requestId'];
			this.pendingRequests.get(requestId)?.resolveFuture(responseParsed);
		});
	}

	initialize(): void {
		this.networkAdapter.initialize();
	}

	teardown(): void {
		this.networkAdapter.teardown();
	}

	callFunctions(...functions: PropertyFunc[]): void {

	}

	setProperties(...properties: Property[]): Promise<Invalidation> {
		this.networkAdapter.publish(JSON.stringify(properties));
		return new Promise<Invalidation>((resolve, reject) => {
			const future = new Future((responseRaw) => resolve(JSON.parse(responseRaw)), reject);
			this.pendingRequests.set(uuidv4(), future);
		});
	}

	setPropertiesAndCallFunctions(properties: Property[], functions: PropertyFunc[]): void {

	}
}

class Future {

	private readonly resolve: (responseRaw: string) => void;
	private readonly reject: () => void;

	constructor(resolve: (responseRaw: string) => void, reject: () => void) {
		this.resolve = resolve;
		this.reject = reject;
	}

	resolveFuture(responseRaw: string): void {
		this.resolve(responseRaw);
	}
}
