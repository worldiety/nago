import type { Invalidation } from '@/shared/model/invalidation';
import type NetworkAdapter from '@/shared/network/networkAdapter';
import type { Property } from '@/shared/model/property';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import { v4 as uuidv4 } from 'uuid';
import type { CallBatch } from '@/shared/network/callBatch';
import type { SetServerProperty } from '@/shared/model/setServerProperty';
import type { CallServerFunc } from '@/shared/model/callServerFunc';
import { useAuth } from '@/stores/authStore';
import type { ClientHello } from '@/shared/network/clientHello';

export default class NetworkProtocol {

	private networkAdapter: NetworkAdapter;
	private pendingFutures: Map<string, Future>;

	constructor(networkAdapter: NetworkAdapter) {
		this.networkAdapter = networkAdapter;
		this.pendingFutures = new Map<string, Future>();
	}

	async initialize(): Promise<Invalidation> {
		await this.networkAdapter.initialize();

		this.networkAdapter.subscribe((responseRaw) => {
			const responseParsed = JSON.parse(responseRaw);
			// TODO: Currently fails, because requestId is not implemented von Go side yet
			const requestId = responseParsed['requestId'];
			this.pendingFutures.get(requestId)?.resolveFuture(responseParsed);
		});

		return this.sendHello();
	}

	private async sendHello(): Promise<Invalidation> {
		const auth = useAuth();

		const hello: ClientHello = {
			type: 'hello',
			auth: {
				keycloak: `${auth.user?.access_token}`,
			},
		};
		const callBatch: CallBatch = {
			tx: [hello],
		};
		return this.publishToAdapter(callBatch);
	}

	teardown(): void {
		this.networkAdapter.teardown();
	}

	async callFunctions(...functions: PropertyFunc[]): Promise<Invalidation|void> {
		const callBatch = this.createCallBatch(undefined, functions);
		if (callBatch.tx.length === 0) {
			return;
		}
		return this.publishToAdapter(callBatch);
	}

	async setProperties(...properties: Property[]): Promise<Invalidation|void> {
		const callBatch = this.createCallBatch(properties);
		if (callBatch.tx.length === 0) {
			return;
		}
		return this.publishToAdapter(callBatch);
	}

	async setPropertiesAndCallFunctions(properties: Property[], functions: PropertyFunc[]): Promise<Invalidation|void> {
		const callBatch = this.createCallBatch(properties, functions);
		if (callBatch.tx.length === 0) {
			return;
		}
		return this.publishToAdapter(callBatch);
	}

	private createCallBatch(properties?: Property[], functions?: PropertyFunc[]): CallBatch {
		const callBatch: CallBatch = {
			tx: [],
		};

		properties
			?.filter((property: Property) => property.id !== 0)
			.forEach((property: Property) => {
				const action: SetServerProperty = {
					type: 'setProp',
					id: property.id,
					value: property.value,
				};
				callBatch.tx.push(action);
			});

		functions
			?.filter((propertyFunc: PropertyFunc) => propertyFunc.id !== 0 && propertyFunc.value !== 0)
			.forEach((propertyFunc: PropertyFunc) => {
				const callServerFunc: CallServerFunc = {
					type: 'callFn',
					id: propertyFunc.value,
				};
				callBatch.tx.push(callServerFunc);
			});

		return callBatch;
	}

	private async publishToAdapter(callBatch: CallBatch): Promise<Invalidation> {
		this.networkAdapter.publish(JSON.stringify(callBatch));
		return new Promise<Invalidation>((resolve, reject) => {
			const future = new Future((responseRaw) => resolve(JSON.parse(responseRaw)), reject);
			this.addFuture(future);
		});
	}

	private addFuture(future: Future): void {
		// Allow a maximum of 10000 pending futures
		if (this.pendingFutures.size >= 10000) {
			const sortedPendingRequests = new Map<string, Future>([...this.pendingFutures.entries()].sort(comparePendingFutures));
			this.pendingFutures.delete(Object.keys(sortedPendingRequests)[0]);
		}

		this.pendingFutures.set(uuidv4(), future);

		function comparePendingFutures(a: [string, Future], b: [string, Future]): number {
			if (a[1].getTimestamp() > b[1].getTimestamp()) {
				return 1;
			} else if (a[1].getTimestamp() < b[1].getTimestamp()) {
				return -1;
			}
			return 0;
		}
	}
}

class Future {

	private readonly resolve: (responseRaw: string) => void;
	private readonly reject: () => void;
	private readonly timestamp: number;

	constructor(resolve: (responseRaw: string) => void, reject: () => void) {
		this.resolve = resolve;
		this.reject = reject;
		this.timestamp = Date.now();
	}

	resolveFuture(responseRaw: string): void {
		this.resolve(responseRaw);
	}

	getTimestamp(): number {
		return this.timestamp;
	}
}
