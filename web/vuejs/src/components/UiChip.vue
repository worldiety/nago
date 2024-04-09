<script lang="ts" setup xmlns="http://www.w3.org/1999/html">
import { computed } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveChip } from '@/shared/model/liveChip';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveChip;
	page: LivePage;
}>();

const networkStore = useNetworkStore();

function onClick() {
	networkStore.invokeFunctions(props.ui.action);
}

function onClose() {
	networkStore.invokeFunctions(props.ui.onClose);
}

const outerColor = computed<string>(() => {
	let cursor = '';
	if (props.ui.action.value != 0) {
		cursor = 'cursor-pointer ';
	}

	switch (props.ui.color.value) {
		case 'red':
			return cursor + 'text-red-800 bg-red-100  dark:bg-red-900 dark:text-red-300';
		case 'green':
			return cursor + 'text-green-800 bg-green-100  dark:bg-green-900 dark:text-green-300';
		case 'yellow':
			return cursor + 'text-yellow-800 bg-yellow-100 dark:bg-yellow-900 dark:text-yellow-300';
		default:
			return cursor + 'text-gray-800 bg-gray-100 dark:bg-gray-700 dark:text-gray-300';
	}
});

const innerColor = computed<string>(() => {
	switch (props.ui.color.value) {
		case 'red':
			return 'hover:bg-red-200 hover:text-red-900 dark:hover:bg-red-800 dark:hover:text-red-300 text-red-400';
		case 'green':
			return 'hover:bg-green-200 hover:text-green-900 dark:hover:bg-green-800 dark:hover:text-green-300';
		case 'yellow':
			return 'hover:bg-yellow-200 hover:text-yellow-900 dark:hover:bg-yellow-800 dark:hover:text-yellow-300';
		default:
			return 'hover:bg-gray-200 hover:text-gray-900 dark:hover:bg-gray-600 dark:hover:text-gray-300';
	}
});
</script>

<template>
	<span
		@click="onClick"
		:class="outerColor"
		class="me-2 inline-flex items-center rounded px-2 py-1 text-sm font-medium"
	>
		{{ props.ui.caption.value }}
		<button
			v-if="props.ui.onClose.value"
			type="button"
			@click="onClose"
			:class="innerColor"
			class="ms-2 inline-flex items-center rounded-sm bg-transparent p-1 text-sm"
			aria-label="Remove"
		>
			<svg class="h-2 w-2" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 14">
				<path
					stroke="currentColor"
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
				/>
			</svg>
		</button>
	</span>
</template>
