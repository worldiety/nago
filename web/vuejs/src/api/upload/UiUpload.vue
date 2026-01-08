<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue';
import type { UploadProgressItem } from '@/api/upload/uploadProgressManager';
import { uploadProgressManager } from '@/api/upload/uploadProgressManager';

const uploads = ref<UploadProgressItem[]>([]);
onMounted(() => {
	const unsubscribe = uploadProgressManager.subscribe((list) => {
		uploads.value = list;
	});
	onUnmounted(unsubscribe);

	//uploadProgressManager.addUpload("1234", "mock.test", 234534)
	//uploadProgressManager.updateProgress("1234", 12)
});

function formatBytes(bytes: number, decimals = 1): string {
	if (bytes === 0) return '0 B';

	const k = 1024;
	const sizes = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];

	const i = Math.floor(Math.log(bytes) / Math.log(k));
	const value = bytes / Math.pow(k, i);

	return `${parseFloat(value.toFixed(decimals))} ${sizes[i]}`;
}
</script>

<template>
	<div v-if="uploads.length > 0" id="uploads" class="inset-0" style="--modal-z-index: 50">
		<div v-for="item in uploads" :key="item.id" class="flex flex-col gap-1 fixed bottom-2 w-full items-center">
			<div class="flex items-center gap-4 p-3 bg-M2 rounded-xl shadow-sm" style="min-width: 30rem">
				<!-- Icon -->
				<div class="">
					<svg
						class="w-6 h-6 text-M8"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linejoin="round"
							stroke-width="2"
							d="M10 3v4a1 1 0 0 1-1 1H5m14-4v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V7.914a1 1 0 0 1 .293-.707l3.914-3.914A1 1 0 0 1 9.914 3H18a1 1 0 0 1 1 1Z"
						/>
					</svg>
				</div>

				<!-- Info -->
				<div class="flex-1 space-y-1 overflow-hidden">
					<div class="flex items-center text-sm text-M8">
						<span class="font-medium truncate">{{ item.fileName }}</span>
						<span class="text-xs text-M5 p-2">|</span>
						<span class="text-xs text-M5">Größe: {{ formatBytes(item.total) }}</span>
						<div class="grow shrink"></div>
						<span class="text-xs text-M5 pl-2">{{ item.progress }}%</span>
					</div>

					<div></div>

					<!-- Progress bar -->
					<div class="w-full bg-M6 h-1.5 rounded-full overflow-hidden mt-1">
						<div
							class="h-full bg-I0 transition-all duration-300"
							:style="{ width: item.progress + '%' }"
						></div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
