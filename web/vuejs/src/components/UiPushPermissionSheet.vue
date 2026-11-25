<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
/*
 * UiPushPermissionSheet asks once more, if the user has asked the application for notifications, but the browser
 * would not ask anymore, because the user gesture has expired, until the call of the backend arrived. Its button is a
 * new user gesture, within which the permission request runs synchronously.
 */
import { useI18n } from 'vue-i18n';
import { pushPermissionRequest } from '@/shared/pushPermissionSheet';

const { t } = useI18n();

function allow() {
	const request = pushPermissionRequest.value;
	pushPermissionRequest.value = null;
	request?.allow();
}

function cancel() {
	const request = pushPermissionRequest.value;
	pushPermissionRequest.value = null;
	request?.cancel();
}
</script>

<template>
	<div
		v-if="pushPermissionRequest"
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="nago-push-permission-title"
	>
		<div class="relative bg-M1 rounded-xl shadow-lg p-6 max-w-sm w-full flex flex-col gap-y-4">
			<h2 id="nago-push-permission-title" class="font-title-medium">{{ t('pushPermission.title') }}</h2>
			<p class="font-body-medium">{{ t('pushPermission.info') }}</p>
			<div class="flex justify-end gap-x-2">
				<button class="button-secondary" @click="cancel">{{ t('pushPermission.cancel') }}</button>
				<button class="button-primary" @click="allow">{{ t('pushPermission.confirm') }}</button>
			</div>
		</div>
	</div>
</template>
