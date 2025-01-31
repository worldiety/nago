import { v4 as uuidv4 } from 'uuid';
import type EventBus from '@/shared/eventbus/eventBus';
import { EventType } from '@/shared/eventbus/eventType';
import ConnectionHandler from '@/shared/network/connectionHandler';
import type ServiceAdapter from '@/shared/network/serviceAdapter';
import { BinaryReader, BinaryWriter, NagoEvent, Ping, marshal, unmarshal } from '@/shared/proto/nprotoc_gen';
import type { Acknowledged } from '@/shared/protocol/ora/acknowledged';
import type { ComponentDestructionRequested } from '@/shared/protocol/ora/componentDestructionRequested';
import type { ComponentFactoryId } from '@/shared/protocol/ora/componentFactoryId';
import type { ComponentInvalidated } from '@/shared/protocol/ora/componentInvalidated';
import type { ConfigurationDefined } from '@/shared/protocol/ora/configurationDefined';
import type { ConfigurationRequested } from '@/shared/protocol/ora/configurationRequested';
import type { Event } from '@/shared/protocol/ora/event';
import type { EventsAggregated } from '@/shared/protocol/ora/eventsAggregated';
import type { FunctionCallRequested } from '@/shared/protocol/ora/functionCallRequested';
import type { NewComponentRequested } from '@/shared/protocol/ora/newComponentRequested';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { ScopeID } from '@/shared/protocol/ora/scopeID';
import type { SetPropertyValueRequested } from '@/shared/protocol/ora/setPropertyValueRequested';
import type { WindowInfo } from '@/shared/protocol/ora/windowInfo';
import type { WindowInfoChanged } from '@/shared/protocol/ora/windowInfoChanged';

export default class WebSocketAdapter implements ServiceAdapter {
	eventBus: EventBus;
	private pendingFutures: Map<number, Future>;
	private destroyedComponents: Map<Ptr, object | null>;
	private readonly webSocketPort: string;
	private readonly isSecure: boolean = false;
	private readonly scopeId: string;
	private webSocket: WebSocket | null = null;
	private closedGracefully: boolean = false;
	private retryTimeout: number | null = null;
	private retries: number = 0;
	private activeLocale: string;
	private requestId: number;
	private bufferCache: Map<Ptr, string>; // TODO reset me, if new component (== new scope) is requested
	private httpBaseUrl: string;

	constructor(eventBus: EventBus) {
		this.eventBus = eventBus;
		this.pendingFutures = new Map();
		this.destroyedComponents = new Map();
		this.bufferCache = new Map<Ptr, string>();
		this.webSocketPort = this.initializeWebSocketPort();
		// important: keep this scopeId for the resume capability only once per
		// channel. Otherwise, (e.g. when storing in localstorage or cookie) all
		// browser tabs and windows will try to steal the scope from each other.
		// So, you MUST ensure that each VueJS instance has its own unique scope id,
		// also when reconnecting to an existing scope.
		this.scopeId = uuidv4();
		this.isSecure = location.protocol == 'https:';
		this.activeLocale = '';
		this.requestId = 0;
	}

	private initializeWebSocketPort(): string {
		let port = import.meta.env.VITE_WS_BACKEND_PORT;
		if (port === '') {
			port = window.location.port;
		}
		return port;
	}

	async initialize(): Promise<void> {
		let proto = 'ws';
		let httpProto = 'http';
		if (this.isSecure) {
			proto = 'wss';
			httpProto = 'https';
		}
		let webSocketURL = `${proto}://${window.location.hostname}:${this.webSocketPort}/wire?_sid=${this.scopeId}`;
		const queryString = window.location.search.substring(1);
		if (queryString) {
			webSocketURL += `&${queryString}`;
		}

		this.httpBaseUrl = `${proto}://${window.location.hostname}:${this.webSocketPort}/`;

		return new Promise<void>((resolve) => {
			this.webSocket = new WebSocket(webSocketURL);
			this.webSocket.binaryType = 'arraybuffer';

			this.webSocket.onmessage = (e) => this.receiveBinary(e.data);

			this.webSocket.onclose = (evt) => {
				ConnectionHandler.connectionChanged({ connected: false });

				if (this.closedGracefully) {
					// Try to reopen the socket if it was not closed gracefully
					window.console.log('ws was intentionally closed');
					this.retry();
				} else {
					// Keep the socket closed if it was closed gracefully (i.e. intentional)
					window.console.log('ws was not closed gracefully');
					console.log('WebSocket closed:', evt);
					console.log('Code:', evt.code);
					console.log('Reason:', evt.reason);
					console.log('Was clean:', evt.wasClean);

					this.closedGracefully = false;
				}
			};

			this.webSocket.onerror = (ev) => {
				window.console.log('websocket failed', ev);
			};

			this.webSocket.onopen = () => {
				ConnectionHandler.connectionChanged({ connected: true });
				this.retries = 0;

				// this keeps our connection at least logically alive
				setInterval(() => {
					if (this.closedGracefully) {
						return;
					}

					this.sendEvent(new Ping());
				}, 30000);

				resolve();
			};
		});
	}

	sendEvent(evt: NagoEvent): void {
		console.log('SEND EVENT', evt);
		let writer = new BinaryWriter();
		marshal(writer, evt);
		let buffer = writer.getBuffer();
		console.log('nprotoc buffer', buffer);
		this.webSocket?.send(buffer);
	}

	async teardown(): Promise<void> {
		window.console.log('websocket teardown');
		this.closedGracefully = true;
		this.webSocket?.close();
		ConnectionHandler.connectionChanged({ connected: false });
	}

	private retry() {
		if (this.retryTimeout !== null) {
			return;
		}

		let timeout = 50; // Retry timeout 50ms
		if (this.retries >= 20) {
			timeout = 250; // 250ms after 2s
		}
		if (this.retries >= 32) {
			timeout = 1000; // 1s after 5s
		}
		if (this.retries >= 37) {
			timeout = 2000; // 2s after 10s
		}

		this.retries += 1;
		this.retryTimeout = window.setTimeout(() => {
			this.retryTimeout = null;
			this.initialize();
		}, timeout);
	}

	async executeFunctions(...functions: Ptr[]): Promise<ComponentInvalidated> {
		return this.send(undefined, functions).then((event) => event as ComponentInvalidated);
	}

	async setProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated> {
		return this.send(properties).then((event) => event as ComponentInvalidated);
	}

	async setPropertiesAndCallFunctions(
		properties: Property<unknown>[],
		functions: Property<Ptr>[]
	): Promise<ComponentInvalidated> {
		return this.send(properties, functions).then((event) => event as ComponentInvalidated);
	}

	getScopeID(): ScopeID {
		return this.scopeId;
	}

	getBufferFromCache(ptr: Ptr): string | undefined {
		return this.bufferCache.get(ptr);
	}

	setBufferToCache(ptr: Ptr, data: string): void {
		this.bufferCache.set(ptr, data);
	}

	async createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated> {
		if (this.activeLocale == '') {
			window.console.log('there is no configured active locale. Invoke getConfiguration to set it.');
		}

		throw 'wer ist das hier?';

		const newComponentRequested: NewComponentRequested = {
			type: 'NewComponentRequested',
			r: this.nextRequestId(), // TODO: Redundant, remove
			activeLocale: this.activeLocale,
			factory: fid,
			values: params,
		};

		return this.send(undefined, undefined, undefined, newComponentRequested).then(
			(event) => event as ComponentInvalidated
		);
	}

	async destroyComponent(ptr: Ptr): Promise<Acknowledged> {
		const componentDestructionRequested: ComponentDestructionRequested = {
			type: 'ComponentDestructionRequested',
			r: this.nextRequestId(), // TODO: Redundant, remove
			ptr: ptr,
		};

		//console.log("async destroy component",ptr)
		this.destroyedComponents.set(ptr, null);

		return this.send(undefined, undefined, undefined, undefined, componentDestructionRequested).then(
			(event) => event as Acknowledged
		);
	}

	async getConfiguration(): Promise<ConfigurationDefined> {
		const winfo: WindowInfo = {
			width: window.innerWidth,
			height: window.innerHeight,
			density: window.devicePixelRatio,
		};

		const configurationRequested: ConfigurationRequested = {
			type: 'ConfigurationRequested',
			r: this.nextRequestId(), // TODO: Redundant, remove
			acceptLanguage: 'de',
			colorScheme: 'default',
			windowInfo: winfo,
		};

		return this.send(undefined, undefined, configurationRequested).then((event) => {
			const configurationDefined = event as ConfigurationDefined;
			this.activeLocale = configurationDefined.activeLocale;
			return configurationDefined;
		});
	}

	private send(
		properties?: Property<unknown>[],
		functions?: Ptr[],
		configurationRequested?: ConfigurationRequested,
		newComponentRequested?: NewComponentRequested,
		componentDestructionRequested?: ComponentDestructionRequested
	): Promise<Event> {
		if (properties?.length > 0) {
			if (properties?.at(0).p == undefined) {
				throw 'fix me';
			}
		}

		if (functions?.length > 0) {
			if (functions?.at(0) == 0) {
				throw 'fix me';
			}
		}

		console.log('shall send', properties);
		return new Promise<Event>((resolve, reject) => {
			const requestId = this.nextRequestId();
			const future = new Future(this, requestId, resolve, reject);
			this.addFuture(future);
			const callBatch = this.createCallBatch(
				requestId,
				properties,
				functions,
				configurationRequested,
				newComponentRequested,
				componentDestructionRequested
			);
			this.webSocket?.send(JSON.stringify(callBatch));
		});
	}

	private receiveBinary(responseRaw: ArrayBuffer): void {
		console.log('WS received', responseRaw);
		let msg = unmarshal(new BinaryReader(new Uint8Array(responseRaw))); // TODO i don't know what i'm doing here, does it copy?
		ConnectionHandler.publishEvent(msg as NagoEvent);
	}

	private receive(responseRaw: string): void {
		const responseParsed = JSON.parse(responseRaw);
		const requestId = responseParsed['r'] as number;
		const event = responseParsed as Event;
		const eventType = event.type as EventType;

		console.log(responseParsed);

		if (eventType == EventType.ERROR_OCCURRED) {
			// this happens, if the server has lost the scope state, either due to a restart or connection timeout
			// TODO this string comparision is stupid and must be changed to at least a numeric error code
			if ((event.message as string).indexOf('no view allocated')) {
				this.eventBus.publish(EventType.ServerStateLost, EventType.PING);
				return;
			}
		}

		// our lowest id is 1, so this must be something without our intention
		if (requestId === 0 || requestId === undefined) {
			//console.log("received unmatched",event.type)
			// something event driven from the backend happened, usually an invalidate or a navigation request
			if (eventType == EventType.INVALIDATED) {
				// it looks like we have a message interleaving problem, where we receive a component tree
				// which is destroyed. The backend cannot send invalidation events of already destroyed components,
				// thus it looks like a logical race at message layer.
				const invalidated = event as ComponentInvalidated;
				//console.log("component invalidated",invalidated.value.id)

				// TODO not sure if this can still be the case
				// if (this.destroyedComponents.has(invalidated.value.id)) {
				// 	console.log("component invalidated illegal",invalidated.value.id)
				// 	return
				// }
			}
			this.eventBus.publish(eventType, event);
			return;
		}

		if (eventType === EventType.ACKNOWLEDGED) {
			// TODO: ack is always the last message of a transaction, however there may be also errors with or without ack.
		}

		this.resolveFuture(requestId, responseParsed);
	}

	private createCallBatch(
		requestId: number,
		properties?: Property<unknown>[],
		functions?: Ptr[],
		configurationRequested?: ConfigurationRequested,
		newComponentRequested?: NewComponentRequested,
		componentDestructionRequested?: ComponentDestructionRequested
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
					v: property.v,
					r: requestId,
				};
				callBatch.events.push(action);
			});

		functions
			// we may be undefined, because the ora protocol is now allowed to omit zero property pointer and values due to performance problems
			?.filter((propertyFunc: Ptr) => propertyFunc != undefined && propertyFunc !== 0)
			.forEach((propertyFunc: Ptr) => {
				const callServerFunc: FunctionCallRequested = {
					type: 'F',
					p: propertyFunc,
					r: requestId,
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
			window.console.log(`error: got network response with unmatched request ID r=${requestId}`);

			const eventType = response.type as EventType;
			this.eventBus.publish(eventType, response);
		} else {
			this.pendingFutures.delete(requestId);
			future.resolveFuture(response);
		}
	}

	private nextRequestId(): number {
		this.requestId++;
		return this.requestId;
	}

	updateWindowInfo(windowInfo: WindowInfo): void {
		const infoChanged: WindowInfoChanged = {
			type: EventType.WindowInfoChanged,
			info: windowInfo,
		};
		this.webSocket?.send(JSON.stringify(infoChanged));
	}
}

class Future {
	private readonly parent: WebSocketAdapter;
	private readonly resolve: (event: Event) => void;
	private readonly reject: (event: Event) => void;
	private readonly monotonicRequestId: number;

	constructor(
		parent: WebSocketAdapter,
		monotonicRequestId: number,
		resolve: (event: Event) => void,
		reject: (event: Event) => void
	) {
		this.resolve = resolve;
		this.reject = reject;
		this.parent = parent;
		this.monotonicRequestId = monotonicRequestId;
	}

	resolveFuture(event: Event): void {
		if (event.type === 'ErrorOccurred') {
			window.console.log(`future ${this.monotonicRequestId} is rejected`);

			// this happens, if the server has lost the scope state, either due to a restart or connection timeout
			// TODO this string comparision is stupid and must be changed to at least a numeric error code
			if ((event.message as string).indexOf('no view allocated')) {
				this.parent.eventBus.publish(EventType.ServerStateLost, EventType.PING);
			}

			this.reject(event);
			return;
		}

		this.resolve(event);
	}

	getRequestId(): number {
		return this.monotonicRequestId;
	}
}
