import type NetworkAdapter from '@/shared/network/networkAdapter';

export default class WebSocketAdapter implements NetworkAdapter {

	private readonly webSocketPort: string;
	private webSocket: WebSocket|null = null;
	private closedGracefully: boolean = false;
	private retryTimeout: number|null = null;
	private scopeId: string;

	constructor() {
		this.webSocketPort = this.initializeWebSocketPort();
		// important: keep this scopeId for the resume capability only once per
		// channel. Otherwise, (e.g. when storing in localstorage or cookie) all
		// browser tabs and windows will try to steal the scope from each other.
		// So, you MUST ensure that each VueJS instance has its own unique scope id,
		// also when reconnecting to an existing scope.
		this.scopeId = window.crypto.randomUUID()
	}

	private initializeWebSocketPort(): string {
		let port = import.meta.env.VITE_WS_BACKEND_PORT;
		if (port === "") {
			port = window.location.port;
		}
		return port;
	}

	async initialize(): Promise<void> {

		let webSocketURL = `ws://${window.location.hostname}:${this.webSocketPort}/wire?_sid=${this.scopeId}`;
		const queryString = window.location.search.substring(1);
		if (queryString) {
			webSocketURL += `&${queryString}`;
		}


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
		if (!this.webSocket){
			console.log("webSocketAdapter is invalid")
		}
		console.log("webSocketAdapter send",payloadRaw)
		this.webSocket?.send(payloadRaw);
	}

	subscribe(resolve: (responseRaw: string) => void): void {
		if (!this.webSocket) {
			console.log("webSocketAdapter rejected subscriber")
			return;
		}
		this.webSocket.onmessage = (e) => resolve(e.data);
		this.webSocket.onerror = this.retry;
	}
}
