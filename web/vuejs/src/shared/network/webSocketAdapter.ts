import type ServiceAdapter from '@/shared/network/serviceAdapter';
import type { Ping } from '@/shared/protocol/gen/ping';
import type { Property } from '@/shared/protocol/property';
import type { Pointer } from '@/shared/protocol/pointer';
import type { Event } from '@/shared/protocol/gen/event';
import type { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import type { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';
import type { ConfigurationRequested } from '@/shared/protocol/gen/configurationRequested';
import type { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import type { EventsAggregated } from '@/shared/protocol/gen/eventsAggregated';
import type { SetPropertyValueRequested } from '@/shared/protocol/gen/setPropertyValueRequested';
import type { FunctionCallRequested } from '@/shared/protocol/gen/functionCallRequested';
import type { NewComponentRequested } from '@/shared/protocol/gen/newComponentRequested';
import type { ComponentDestructionRequested } from '@/shared/protocol/gen/componentDestructionRequested';
import type { ColorScheme } from '@/shared/protocol/colorScheme';
import type { ComponentFactoryId } from '@/shared/protocol/componentFactoryId';
import { v4 as uuidv4 } from 'uuid';
import type EventBus from '@/shared/eventbus/eventBus';
import { EventType } from '@/shared/eventbus/eventType';

export default class WebSocketAdapter implements ServiceAdapter {

	private eventBus: EventBus;
	private pendingFutures: Map<number, Future>;
	private readonly webSocketPort: string;
	private readonly isSecure: boolean = false;
	private readonly scopeId: string;
	private webSocket: WebSocket|null = null;
	private closedGracefully: boolean = false;
	private retryTimeout: number|null = null;
	private activeLocale: string;
	private requestId: number;

	constructor(eventBus: EventBus) {
		this.eventBus = eventBus;
		this.pendingFutures = new Map();
		this.webSocketPort = this.initializeWebSocketPort();
		// important: keep this scopeId for the resume capability only once per
		// channel. Otherwise, (e.g. when storing in localstorage or cookie) all
		// browser tabs and windows will try to steal the scope from each other.
		// So, you MUST ensure that each VueJS instance has its own unique scope id,
		// also when reconnecting to an existing scope.
		this.scopeId = uuidv4();
		this.isSecure = location.protocol == "https:";
		this.activeLocale = '';
		this.requestId = 0;
	}

	private initializeWebSocketPort(): string {
		let port = import.meta.env.VITE_WS_BACKEND_PORT;
		if (port === "") {
			port = window.location.port;
		}
		return port;
	}

	async initialize(): Promise<void> {
		let proto = "ws";
		if (this.isSecure) {
			proto = "wss";
		}
		let webSocketURL = `${proto}://${window.location.hostname}:${this.webSocketPort}/wire?_sid=${this.scopeId}`;
		const queryString = window.location.search.substring(1);
		if (queryString) {
			webSocketURL += `&${queryString}`;
		}

		return new Promise<void>((resolve) => {
			this.webSocket = new WebSocket(webSocketURL);

			this.webSocket.onmessage = (e) => this.receive(e.data);

			this.webSocket.onclose = () => {
				if (!this.closedGracefully) {
					// Try to reopen the socket if it was not closed gracefully
					this.retry();
				} else {
					// Keep the socket closed if it was closed gracefully (i.e. intentional)
					this.closedGracefully = false;
				}
			}

			this.webSocket.onopen = () => {
				// this keeps our connection at least logically alive
				setInterval(()=>{
					if (this.closedGracefully){
						return
					}
					const evt: Ping = {
						type: 'Ping',
					};

					this.webSocket?.send(JSON.stringify(evt))
				},30000);

				resolve();
			}
		})
	}

	async teardown(): Promise<void> {
		this.closedGracefully = true;
		this.webSocket?.close();
	}

	private retry() {
		if (this.retryTimeout !== null) {
			return;
		}
		this.retryTimeout = window.setTimeout(() => {
			this.retryTimeout = null;
			this.initialize();
		}, 2000);
	}

	async executeFunctions(...functions: Property<Pointer>[]): Promise<ComponentInvalidated> {
		return this.send(undefined, functions).then((event) => event as ComponentInvalidated);
	}

	async setProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated> {
		return this.send(properties).then((event) => event as ComponentInvalidated);
	}

	async setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated> {
		return this.send(properties, functions).then((event) => event as ComponentInvalidated);
	}

	async createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated> {
		if (this.activeLocale == "") {
			console.log("there is no configured active locale. Invoke getConfiguration to set it.")
		}

		const newComponentRequested: NewComponentRequested = {
			type: 'NewComponentRequested',
			r: this.nextRequestId(), // TODO: Redundant, remove
			activeLocale: this.activeLocale,
			factory: fid,
			values: params,
		};

		return this.send(
			undefined,
			undefined,
			undefined,
			newComponentRequested,
		).then((event) => event as ComponentInvalidated);
	}

	async destroyComponent(ptr: Pointer): Promise<Acknowledged> {
		const componentDestructionRequested: ComponentDestructionRequested = {
			type: 'ComponentDestructionRequested',
			r: this.nextRequestId(), // TODO: Redundant, remove
			ptr: ptr,
		};

		return this.send(
			undefined,
			undefined,
			undefined,
			undefined,
			componentDestructionRequested,
		).then((event) => event as Acknowledged);
	}

	async getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined> {
		const configurationRequested: ConfigurationRequested = {
			type: 'ConfigurationRequested',
			r: this.nextRequestId(), // TODO: Redundant, remove
			acceptLanguage: acceptLanguages,
			colorScheme: colorScheme,
		};

		return this.send(undefined, undefined, configurationRequested).then((event) => {
			const configurationDefined = event as ConfigurationDefined;
			this.activeLocale = configurationDefined.activeLocale;
			return configurationDefined;
		});
	}

	private send(
		properties?: Property<unknown>[],
		functions?: Property<Pointer>[],
		configurationRequested?: ConfigurationRequested,
		newComponentRequested?: NewComponentRequested,
		componentDestructionRequested?: ComponentDestructionRequested,
	): Promise<Event> {
		return new Promise<Event>((resolve, reject) => {
			const requestId = this.nextRequestId();
			const future = new Future(requestId, resolve, reject);
			this.addFuture(future);
			const callBatch = this.createCallBatch(requestId, properties, functions, configurationRequested, newComponentRequested, componentDestructionRequested);
			this.webSocket?.send(JSON.stringify(callBatch));
		});
	}

	private receive(responseRaw: string): void {
		const responseParsed = JSON.parse(responseRaw);
		const requestId = responseParsed['r'] as number;
		const event = responseParsed as Event;
		const eventType = event.type as EventType;

		// our lowest id is 1, so this must be something without our intention
		if (requestId === 0 || requestId === undefined) {
			// something event driven from the backend happened, usually an invalidate or a navigation request
			this.eventBus.publish(eventType, event);
			return;
		}

		if (eventType === EventType.ACKNOWLEDGED) {
			// TODO: Backend needs to tell the frontend about the end of a transaction (EOT / amount of messages / ...)
		}

		this.resolveFuture(requestId, responseParsed);
	}

	private createCallBatch(
		requestId: number,
		properties?: Property<unknown>[],
		functions?: Property<Pointer>[],
		configurationRequested?: ConfigurationRequested,
		newComponentRequested?: NewComponentRequested,
		componentDestructionRequested?: ComponentDestructionRequested,
	): EventsAggregated {
		const callBatch: EventsAggregated = {
			type: 'T',
			events: [],
			r: requestId,
		};

		properties
			?.filter((property: Property<unknown>) => property.p !== 0)
			.forEach((property: Property<unknown>) => {
				const action: SetPropertyValueRequested = {
					type: 'P',
					p: property.p,
					v: String(property.v),
					r: requestId, // TODO: Redundant, remove
				};
				callBatch.events.push(action);
			});

		functions
			?.filter((propertyFunc: Property<Pointer>) => propertyFunc.p !== 0 && propertyFunc.v !== 0)
			.forEach((propertyFunc: Property<Pointer>) => {
				const callServerFunc: FunctionCallRequested = {
					type: 'F',
					p: propertyFunc.v,
					r: requestId, // TODO: Redundant, remove
				};
				callBatch.events.push(callServerFunc);
			});

		if (configurationRequested) {
			configurationRequested.r = requestId;
			callBatch.events.push(configurationRequested);
		}

		if (newComponentRequested) {
			newComponentRequested.r = requestId;
			callBatch.events.push(newComponentRequested);
		}

		if (componentDestructionRequested) {
			componentDestructionRequested.r = requestId;
			callBatch.events.push(componentDestructionRequested);
		}

		return callBatch;
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

	private resolveFuture(requestId: number, response: Event): void {
		const future = this.pendingFutures.get(requestId);
		if (!future) {
			console.log(`error: got network response with unmatched request ID r=${requestId}`)
		} else {
			this.pendingFutures.delete(requestId)
			future.resolveFuture(response);
		}
	}

	private nextRequestId(): number {
		this.requestId++;
		return this.requestId;
	}
}

class Future {

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
