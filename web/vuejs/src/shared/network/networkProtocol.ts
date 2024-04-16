import type {Invalidation} from '@/shared/model/invalidation';
import type NetworkAdapter from '@/shared/network/networkAdapter';
import {ConfigurationRequested} from "@/shared/protocol/gen/configurationRequested";
import {ColorScheme} from "@/shared/protocol/colorScheme";
import {ConfigurationDefined} from "@/shared/protocol/gen/configurationDefined";
import {ComponentFactoryId} from "@/shared/protocol/componentFactoryId";
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {NewComponentRequested} from "@/shared/protocol/gen/newComponentRequested";
import {Property} from "@/shared/protocol/property";
import {Pointer} from "@/shared/protocol/pointer";
import {EventsAggregated} from "@/shared/protocol/gen/eventsAggregated";
import {SetPropertyValueRequested} from "@/shared/protocol/gen/setPropertyValueRequested";
import {FunctionCallRequested} from "@/shared/protocol/gen/functionCallRequested";
import {Event} from "@/shared/protocol/gen/event";

export default class NetworkProtocol {

	private networkAdapter: NetworkAdapter;
	private pendingFutures: Map<number, Future>;
	private reqCounter: number;
	private activeLocale: string;
	private unprocessedEventSubscribers: ((evt: Event) => void)[]

	constructor(networkAdapter: NetworkAdapter) {
		this.networkAdapter = networkAdapter;
		this.pendingFutures = new Map<number, Future>();
		this.reqCounter = 1;
		this.activeLocale = "";
		this.unprocessedEventSubscribers = [];
	}

	async initialize(): Promise<void> {
		await this.networkAdapter.initialize();
		console.log("networkAdapter is ok")

		this.networkAdapter.subscribe((responseRaw) => {
			//console.log("got response", responseRaw)
			const responseParsed = JSON.parse(responseRaw);
			let requestId = responseParsed['requestId'] as number;
			if (requestId === undefined) {
				// try again the shortened field name of ack, we keep that efficient
				requestId = responseParsed['r'] as number;
			}

			console.log(responseParsed)
			// our lowest id is 1, so this must be something without our intention
			if (requestId === 0) {
				// something event driven from the backend happened, usually an invalidate or a navigation request
				console.log(`received unrequested event from backend: ${responseParsed.type}`)
				this.unprocessedEventSubscribers.forEach(fn => {
					if (fn === undefined) {
						return
					}

					fn(responseParsed as Event)
				})

				return
			}

			let future = this.pendingFutures.get(requestId);
			if (!future) {
				console.log(`error: got network response with unmatched requestId=${requestId}`)
			} else {
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

	addUnprocessedEventSubscriber(fn: ((evt: Event) => void)) {
		this.unprocessedEventSubscribers.push(fn)
	}

	removeUnprocessedEventSubscriber(fn: ((evt: Event) => void)) {
		this.unprocessedEventSubscribers = this.unprocessedEventSubscribers.filter(obj => obj !== fn)
	}

	async getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined> {
		const evt: ConfigurationRequested = {
			type: 'ConfigurationRequested',
			requestId: this.nextReqId(),
			acceptLanguage: acceptLanguages,
			colorScheme: colorScheme,
		};

		return this.publishToAdapter(evt.requestId, evt).then(value => {
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

		return this.publishToAdapter(evt.requestId, evt).then(value => value as ComponentInvalidated)
	}


	teardown(): void {
		this.networkAdapter.teardown();
	}

	// todo I don't believe a void choice type is a good idea...
	async callFunctions(...functions: Property<Pointer>[]): Promise<Invalidation | void> {
		let rid = this.nextReqId();
		const callBatch = this.createCallBatch(rid, undefined, functions);
		if (callBatch.events.length === 0) {
			return;
		}

		return this.publishToAdapter(rid, callBatch);
	}

	// todo I don't believe a void choice type is a good idea...
	async setProperties(...properties: Property<unknown>[]): Promise<Invalidation | void> {
		let rid = this.nextReqId();
		const callBatch = this.createCallBatch(rid, properties, undefined);
		if (callBatch.events.length === 0) {
			return;
		}
		return this.publishToAdapter(rid, callBatch);
	}

	// todo I don't believe a void choice type is a good idea...
	async setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<Invalidation | void> {
		let rid = this.nextReqId();
		const callBatch = this.createCallBatch(rid, properties, functions);
		if (callBatch.events.length === 0) {
			return;
		}
		return this.publishToAdapter(rid, callBatch);
	}

	private createCallBatch(requestId: number, properties?: Property<unknown>[], functions?: Property<Pointer>[]): EventsAggregated {
		const callBatch: EventsAggregated = {
			type: "T",
			events: [],
			r: requestId,
		};

		properties
			?.filter((property: Property<unknown>) => property.p !== 0)
			.forEach((property: Property<unknown>) => {
				const action: SetPropertyValueRequested = {
					type: 'P',
					p: property.p,
					v: String(property.v), // TODO is this correct to convert any into a string?
					//requestId: requestId, // TODO logically not required and inefficient due to repetition on call , make me optional
				};
				callBatch.events.push(action);
			});

		functions
			?.filter((propertyFunc: Property<Pointer>) => propertyFunc.p !== 0 && propertyFunc.v !== 0)
			.forEach((propertyFunc: Property<Pointer>) => {
				const callServerFunc: FunctionCallRequested = {
					type: 'F',
					p: propertyFunc.v,
					//requestId: requestId, // TODO logically not required and inefficient due to repetition on call, make me optional
				};
				callBatch.events.push(callServerFunc);
			});

		return callBatch;
	}

	private async publishToAdapter(requestId: number, evt: unknown): Promise<unknown> {
		this.networkAdapter.publish(JSON.stringify(evt));
		return new Promise<unknown>((resolve, reject) => {
			const future = new Future(requestId, (obj) => {
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
	private readonly reject: (reason: unknown) => void;
	private readonly monotonicRequestId: number;

	constructor(monotonicRequestId: number, resolve: (responseRaw: unknown) => void, reject: (reason: unknown) => void) {
		this.resolve = resolve;
		this.reject = reject;
		this.monotonicRequestId = monotonicRequestId;
	}

	resolveFuture(responseRaw: unknown): void {
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
