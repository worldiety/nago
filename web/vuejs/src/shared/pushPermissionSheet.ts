/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

import { ref } from 'vue';

/** PushPermissionRequest is a pending question of UiPushPermissionSheet. */
export interface PushPermissionRequest {
	allow: () => void;
	cancel: () => void;
}

/**
 * pushPermissionRequest is shown by UiPushPermissionSheet. It is only set, if the user has asked for notifications,
 * but the user gesture has expired, until the call of the backend arrived.
 */
export const pushPermissionRequest = ref<PushPermissionRequest | null>(null);

/** confirmPushPermission shows the sheet. A pending question is cancelled. */
export function confirmPushPermission(allow: () => void, cancel: () => void): void {
	pushPermissionRequest.value?.cancel();
	pushPermissionRequest.value = { allow, cancel };
}
