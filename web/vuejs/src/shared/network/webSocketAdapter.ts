import { useAuth, UserChangedCallbacks } from '@/stores/authStore';
import type { UpdateJWT } from '@/shared/model/updateJWT';
import type { CallBatch } from '@/shared/model/callBatch';
import type { ClientHello } from '@/shared/model/clientHello';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { SetServerProperty } from '@/shared/model/setServerProperty';
import type { CallServerFunc } from '@/shared/model/callServerFunc';
import type NetworkAdapter from '@/shared/network/networkAdapter';
import type { Property } from '@/shared/model/property';

export default class WebSocketAdapter implements NetworkAdapter {

	private readonly webSocketPort: string;
	private webSocket: WebSocket|null = null;
	private closedGracefully: boolean = false;
	private retryTimeout: number|null = null;

	constructor() {
		this.webSocketPort = this.initializeWebSocketPort();
	}

	private initializeWebSocketPort(): string {
		let port = import.meta.env.VITE_WS_BACKEND_PORT;
		if (port === "") {
			port = window.location.port;
		}
		return port;
	}

	initialize(): void {
		let webSocketURL = `ws://${window.location.hostname}:${this.webSocketPort}/wire?_pid=${window.location.pathname.substring(1)}`;
		const queryString = window.location.search.substring(1);
		if (queryString) {
			webSocketURL += `&${queryString}`;
		}

		this.webSocket = new WebSocket(webSocketURL);

		this.webSocket.onopen = () => {
			this.sendHello();
		}

		this.webSocket.onclose = () => {
			if (!this.closedGracefully) {
				// Try to reopen the socket if it was not closed gracefully
				this.retry();
			} else {
				// Keep the socket closed if it was closed gracefully (i.e. intentional)
				this.closedGracefully = false;
			}
		}

		UserChangedCallbacks.push(() => this.sendUser());
	}

	private sendUser(){
		const auth = useAuth();

		const updateJWT: UpdateJWT = {
			type: "updateJWT",
			token: `${auth.user?.access_token}`,
			OIDCName:"Keycloak",
		}

		const callTx: CallBatch = {
			tx: [updateJWT]
		}

		this.webSocket?.send(JSON.stringify(callTx))
	}

	private sendHello(){
		const auth = useAuth();

		const hello: ClientHello = {
			type: "hello",
			auth: {
				keycloak: `${auth.user?.access_token}`,
			},
		}

		const callTx: CallBatch = {
			tx: [hello]
		}

		this.webSocket?.send(JSON.stringify(callTx))
	}

	teardown(): void {
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

	publish(payloadRaw: string): void {
		// TODO: Figure out how to call invokeTx2 without knowing the exact parsed type of payloadRaw
	}

	subscribe(resolve: (responseRaw: string) => void): void {
		if (!this.webSocket) {
			return;
		}
		this.webSocket.onmessage = (e) => resolve(e.data);

		this.webSocket.onerror = () => {
			// TODO: Figure out how to handle failed publishes (i.e. no response with a request ID is returned)
			this.retry();
		}
	}

	private invokeTx2(properties?: Property[], functions?: PropertyFunc[]) {
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

		if (callBatch.tx.length > 0) {
			this.webSocket?.send(JSON.stringify(callBatch));
		}
	}
}
