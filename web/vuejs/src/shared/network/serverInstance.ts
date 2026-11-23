/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * ServerInstance reloads the page, when a connection reaches another instance of the server than the first one,
 * e.g. after a deployment, because the frontend may not match anymore.
 */
export default class ServerInstance {
	private id = '';

	check(): void {
		fetch('/api/nago/v1/instance')
			.then((response) => response.json())
			.then((data: { id: string }) => {
				if (this.id === '') {
					this.id = data.id;
				} else if (this.id !== data.id) {
					console.log('Server instance changed, reloading page...');
					window.location.reload();
				}
			})
			.catch((err) => {
				console.error('Failed to fetch instance version:', err);
			});
	}
}
