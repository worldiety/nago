/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import type { ConnectionState } from '@/shared/network/connectionState';
import { NagoEvent } from '@/shared/proto/nprotoc_gen';

export default class ConnectionHandler {
	private static readonly changeListeners: ((connectionState: ConnectionState) => void)[] = [];
	private static readonly eventListeners: ((evt: NagoEvent) => void)[] = [];

	public static addConnectionChangeListener(callback: (connectionState: ConnectionState) => void): void {
		this.changeListeners.push(callback);
	}

	public static connectionChanged(connectionState: ConnectionState): void {
		this.changeListeners.forEach((callback) => callback(connectionState));
	}

	public static addEventListener(callback: (evt: NagoEvent) => void): void {
		this.eventListeners.push(callback);
	}

	public static removeEventListener(callback: (evt: NagoEvent) => void): void {
		const index = this.eventListeners.indexOf(callback);
		if (index !== -1) {
			this.eventListeners.splice(index, 1);
		}
	}

	public static removeConnectionChangeListener(callback: (connectionState: ConnectionState) => void): void {
		const index = this.changeListeners.indexOf(callback);
		if (index !== -1) {
			this.changeListeners.splice(index, 1);
		}
	}

	public static publishEvent(evt: NagoEvent): void {
		this.eventListeners.forEach((callback) => callback(evt));
	}

	// reset removes all registered change or event listeners.
	public static reset(): void {
		this.changeListeners.length = 0;
		this.eventListeners.length = 0;
	}
}
