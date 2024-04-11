import { useAuth, UserChangedCallbacks } from '@/stores/authStore';
import type { UpdateJWT } from '@/shared/model/updateJWT';
import type { CallBatch } from '@/shared/network/callBatch';
import type NetworkAdapter from '@/shared/network/networkAdapter';

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

	async initialize(): Promise<void> {
		let webSocketURL = `ws://${window.location.hostname}:${this.webSocketPort}/wire?_pid=${window.location.pathname.substring(1)}`;
		const queryString = window.location.search.substring(1);
		if (queryString) {
			webSocketURL += `&${queryString}`;
		}

		UserChangedCallbacks.push(() => this.sendUser());

		return new Promise<void>((resolve) => {
			this.webSocket = new WebSocket(webSocketURL);

			this.webSocket.onclose = () => {
				if (!this.closedGracefully) {
					// Try to reopen the socket if it was not closed gracefully
					this.retry();
				} else {
					// Keep the socket closed if it was closed gracefully (i.e. intentional)
					this.closedGracefully = false;
				}
			}

			this.webSocket.onopen = () => resolve();
		})
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
		this.webSocket?.send(payloadRaw);
	}

	subscribe(resolve: (responseRaw: string) => void): void {
		if (!this.webSocket) {
			return;
		}
		this.webSocket.onmessage = (e) => resolve(e.data);
		this.webSocket.onerror = this.retry;
	}
}
