import type {Invalidation} from '@/shared/model/invalidation';
import type NetworkAdapter from '@/shared/network/networkAdapter';
import type {Property} from '@/shared/model/property';
import type {PropertyFunc} from '@/shared/model/propertyFunc';
import {v4 as uuidv4} from 'uuid';
import type {CallBatch} from '@/shared/network/callBatch';
import type {SetServerProperty} from '@/shared/model/setServerProperty';
import type {CallServerFunc} from '@/shared/model/callServerFunc';
import {ConfigurationRequested} from "@/shared/protocol/gen/configurationRequested";
import {ColorScheme} from "@/shared/protocol/colorScheme";
import {ConfigurationDefined} from "@/shared/protocol/gen/configurationDefined";
import {ComponentFactoryId} from "@/shared/protocol/componentFactoryId";
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {NewComponentRequested} from "@/shared/protocol/gen/newComponentRequested";

export default class NetworkProtocol {

	private networkAdapter: NetworkAdapter;
	private pendingFutures: Map<number, Future>;
	private reqCounter: number;
	private activeLocale: string;

	constructor(networkAdapter: NetworkAdapter) {
		this.networkAdapter = networkAdapter;
		this.pendingFutures = new Map<number, Future>();
		this.reqCounter = 1;
		this.activeLocale = "";
	}

	async initialize(): Promise<void> {
		await this.networkAdapter.initialize();
		console.log("networkAdapter is ok")

		this.networkAdapter.subscribe((responseRaw) => {
			console.log("got response",responseRaw)
			const responseParsed = JSON.parse(responseRaw);
			const requestId = responseParsed['requestId'] as number;
			let future = this.pendingFutures.get(requestId);
			if (!future){
				console.log(`error: got network response with unmatched requestId=${requestId}`)
			}else{
				this.pendingFutures.delete(requestId)
				future.resolveFuture(responseParsed);
			}

		});

		return new Promise(resolve => resolve())
	}

	private nextReqId(): number {
		this.reqCounter++;
		return this.reqCounter;
	}

	async getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined> {
		const evt: ConfigurationRequested = {
			type: 'ConfigurationRequested',
			requestId: this.nextReqId(),
			acceptLanguage: acceptLanguages,
			colorScheme: colorScheme,
		};

		return this.publishToAdapter(evt.requestId,evt).then(value => {
			let evt = value as ConfigurationDefined
			this.activeLocale = evt.activeLocale;
			return evt
		});
	}

	async newComponent(fid: ComponentFactoryId, params: Map<string, string>): Promise<ComponentInvalidated> {
		if (this.activeLocale == "") {
			console.log("there is no configured active locale. Invoke getConfiguration to set it.")
		}

		const evt: NewComponentRequested = {
			type: 'NewComponentRequested',
			requestId: this.nextReqId(),
			activeLocale: this.activeLocale,
			factory: fid,
			values: params,
		};

		return this.publishToAdapter(evt.requestId,evt).then(value => value as ComponentInvalidated)
	}


	teardown(): void {
		this.networkAdapter.teardown();
	}

	async callFunctions(...functions: PropertyFunc[]): Promise<Invalidation | void> {
		const callBatch = this.createCallBatch(undefined, functions);
		if (callBatch.tx.length === 0) {
			return;
		}
		return this.publishToAdapter(callBatch);
	}

	async setProperties(...properties: Property[]): Promise<Invalidation | void> {
		const callBatch = this.createCallBatch(properties);
		if (callBatch.tx.length === 0) {
			return;
		}
		return this.publishToAdapter(callBatch);
	}

	async setPropertiesAndCallFunctions(properties: Property[], functions: PropertyFunc[]): Promise<Invalidation | void> {
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

	private async publishToAdapter(requestId:number,evt: unknown): Promise<unknown> {
		this.networkAdapter.publish(JSON.stringify(evt));
		return new Promise<unknown>((resolve, reject) => {
			const future = new Future(requestId,(obj) => {
				//console.log(obj)
				return resolve(obj);
			}, reject);
			this.addFuture(future);
		});
	}

	private addFuture(future: Future): void {
		// Allow a maximum of 10000 pending futures
		if (this.pendingFutures.size >= 10000) {

			const sortedPendingRequests = [...this.pendingFutures.entries()].sort(comparePendingFutures);
			this.pendingFutures.delete(sortedPendingRequests[0][0]);
		}

		this.pendingFutures.set(future.getRequestId(), future);

		function comparePendingFutures(a: [number, Future], b: [number, Future]): number {
			if (a[1].getRequestId() > b[1].getRequestId()) {
				return 1;
			} else if (a[1].getRequestId() < b[1].getRequestId()) {
				return -1;
			}
			return 0;
		}
	}
}

class Future {

	private readonly resolve: (responseRaw: unknown) => void;
	private readonly reject: () => void;
	private readonly monotonicRequestId: number;

	constructor(monotonicRequestId:number,resolve: (responseRaw: unknown) => void, reject: () => void) {
		this.resolve = resolve;
		this.reject = reject;
		this.monotonicRequestId = monotonicRequestId;
	}

	resolveFuture(responseRaw: unknown): void {
		this.resolve(responseRaw);
	}

	getRequestId(): number {
		return this.monotonicRequestId;
	}
}
