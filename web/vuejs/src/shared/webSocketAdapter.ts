import { useAuth, UserChangedCallbacks } from '@/stores/authStore';
import type { LiveMessage } from '@/shared/model/liveMessage';
import type { UpdateJWT } from '@/shared/model/updateJWT';
import type { CallBatch } from '@/shared/model/callBatch';
import type { ClientHello } from '@/shared/model/clientHello';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { Property } from '@/shared/model/property';
import type { SetServerProperty } from '@/shared/model/setServerProperty';
import type { CallServerFunc } from '@/shared/model/callServerFunc';

export default class WebSocketAdapter {

	private readonly webSocketPort: string;
	private webSocket: WebSocket|null = null;
	private webSocketReceiveCallback: ((message: LiveMessage) => void) | null = null;
	private webSocketErrorCallback: (() => void) | null = null;
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

	initializeWebSocket(): void {
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

		this.webSocket.onmessage = (e) => {
			const message: LiveMessage = JSON.parse(e.data)
			if (this.webSocketReceiveCallback) {
				this.webSocketReceiveCallback(message);
			}
		}

		this.webSocket.onerror = (e) => {
			if (this.webSocketErrorCallback) {
				this.webSocketErrorCallback();
			}
			this.retry();
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

	private retry() {
		if (this.retryTimeout !== null) {
			return;
		}
		this.retryTimeout = window.setTimeout(() => {
			this.retryTimeout = null;
			this.initializeWebSocket();
		}, 2000);
	}

	invokeFunctions(...actions: PropertyFunc[]) {
		this.invokeTx2(undefined, actions);
	}

	invokeSetProperties(...properties: Property[]) {
		this.invokeTx2(properties);
	}

	invokeFunctionsAndSetProperties(properties: Property[], functions: PropertyFunc[]) {
		this.invokeTx2(properties, functions);
	}

	private invokeTx2(properties?: Property[], functions?: PropertyFunc[]) {
		const callBatch: CallBatch = {
			tx: [],
		};

		properties
			?.filter((property: Property) => property.id !== 0)
			.forEach((property: Property) => {
				const setServerProperty: SetServerProperty = {
					type: 'setProp',
					id: property.id,
					value: property.value,
				};
				callBatch.tx.push(setServerProperty);
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

	setWebSocketReceiveCallback(callback: (message: LiveMessage) => void): void {
		this.webSocketReceiveCallback = callback;
	}

	setWebSocketErrorCallback(callback: () => void): void {
		this.webSocketErrorCallback = callback;
	}

	closeWebSocket(): void {
		this.closedGracefully = true;
		this.webSocket?.close();
	}
}
