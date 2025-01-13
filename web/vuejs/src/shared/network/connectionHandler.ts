import type { ConnectionState } from '@/shared/network/connectionState';

export default class ConnectionHandler {
	private static readonly changeListeners: ((connectionState: ConnectionState) => void)[] = [];

	public static addConnectionChangeListener(callback: (connectionState: ConnectionState) => void): void {
		this.changeListeners.push(callback);
	}

	public static connectionChanged(connectionState: ConnectionState): void {
		this.changeListeners.forEach((callback) => callback(connectionState));
	}
}
