import NetworkAdapter from '@/shared/network/networkAdapter';
import type { Ping } from "@/shared/protocol/gen/ping";
import { Property } from '@/shared/protocol/property';
import { Pointer } from '@/shared/protocol/pointer';
import { Event } from '@/shared/protocol/gen/event';
import { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import { ColorScheme } from '@/shared/protocol/colorScheme';
import { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';
import { ConfigurationRequested } from '@/shared/protocol/gen/configurationRequested';
import { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import Future from '@/shared/network/future';
import type { EventsAggregated } from '@/shared/protocol/gen/eventsAggregated';
import { SetPropertyValueRequested } from '@/shared/protocol/gen/setPropertyValueRequested';
import { FunctionCallRequested } from '@/shared/protocol/gen/functionCallRequested';

export default class WebSocketAdapter extends NetworkAdapter {

	private readonly webSocketPort: string;
	private readonly isSecure: boolean = false;
	private webSocket: WebSocket|null = null;
	private closedGracefully: boolean = false;
	private retryTimeout: number|null = null;
	private scopeId: string;

	constructor() {
		super();
		this.webSocketPort = this.initializeWebSocketPort();
		// important: keep this scopeId for the resume capability only once per
		// channel. Otherwise, (e.g. when storing in localstorage or cookie) all
		// browser tabs and windows will try to steal the scope from each other.
		// So, you MUST ensure that each VueJS instance has its own unique scope id,
		// also when reconnecting to an existing scope.
		this.scopeId = window.crypto.randomUUID()
		this.isSecure = location.protocol == "https:";
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

			this.webSocket.onmessage = (e) => this.processReceivedMessage(e.data);

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

	private processReceivedMessage(responseRaw: any): void {
		const responseParsed = JSON.parse(responseRaw);
		let requestId = responseParsed['requestId'] as number;
		if (requestId === undefined) {
			// try again the shortened field name of ack, we keep that efficient
			requestId = responseParsed['r'] as number;
		}

		// our lowest id is 1, so this must be something without our intention
		if (requestId === 0 || requestId === undefined) {
			// something event driven from the backend happened, usually an invalidate or a navigation request
			/*console.log(`received unrequested event from backend: ${responseParsed.type}`)
			this.unprocessedEventSubscribers.forEach(fn => {
				if (fn === undefined) {
					return
				}

				fn(responseParsed as Event)
			})

			return*/
			this.handleUnrequestedMessage(responseParsed as Event);
		}

		this.resolveFuture(requestId);
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

	executeFunctions(functions: Property<Pointer>[]): Promise<ComponentInvalidated|void> {
		return new Promise<ComponentInvalidated | void>((resolve, reject) => {
			const requestId = this.nextReqId();
			const future = new Future(requestId, resolve, reject);
			this.addFuture(future);
			const callBatch = this.createCallBatch(requestId, undefined, functions);
			this.send(callBatch);
		});
	}

	setProperties<T>(properties: Property<T>[]): Promise<ComponentInvalidated|void> {
		return new Promise<ComponentInvalidated | void>((resolve, reject) => {
			const requestId = this.nextReqId();
			const future = new Future(requestId, resolve, reject);
			this.addFuture(future);
			const callBatch = this.createCallBatch(requestId, properties);
			this.send(callBatch);
		});
	}

	setPropertiesAndCallFunctions<T>(properties: Property<T>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated|void> {
		return new Promise<ComponentInvalidated | void>((resolve, reject) => {
			const requestId = this.nextReqId();
			const future = new Future(requestId, resolve, reject);
			this.addFuture(future);
			const callBatch = this.createCallBatch(requestId, properties, functions);
			this.send(callBatch);
		});
	}

	createComponent(): Promise<ComponentInvalidated> {
		return Promise.resolve(undefined);
	}

	destroyComponent(pointer: Pointer): Promise<Acknowledged> {
		return Promise.resolve(undefined);
	}

	getConfiguration(configurationRequested: ConfigurationRequested): Promise<ConfigurationDefined> {
		return new Promise<ConfigurationDefined>((resolve, reject) => {
			const requestId = this.nextReqId();
			const future = new Future(requestId, resolve, reject);
			this.addFuture(future);
			const callBatch = this.createCallBatch(requestId, undefined, undefined, configurationRequested);
			this.send(callBatch);
		});
	}

	private createCallBatch(requestId: number, properties?: Property<unknown>[], functions?: Property<Pointer>[], configurationRequested?: ConfigurationRequested): EventsAggregated {
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
					v: property.v as string,
				};
				callBatch.events.push(action);
			});

		functions
			?.filter((propertyFunc: Property<Pointer>) => propertyFunc.p !== 0 && propertyFunc.v !== 0)
			.forEach((propertyFunc: Property<Pointer>) => {
				const callServerFunc: FunctionCallRequested = {
					type: 'F',
					p: propertyFunc.v,
				};
				callBatch.events.push(callServerFunc);
			});

		if (configurationRequested) {
			callBatch.events.push(configurationRequested);
		}

		return callBatch;
	}

	private send(callBatch: EventsAggregated): void {
		this.webSocket?.send(JSON.stringify(callBatch));
	}
}
