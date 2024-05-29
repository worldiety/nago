<script setup lang="ts">
import { computed } from 'vue';
import type {Chip} from "@/shared/protocol/ora/chip";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Chip;
}>();

const serviceAdapter = useServiceAdapter();

function onClick() {
	serviceAdapter.executeFunctions(props.ui.action);
}

function onClose() {
	serviceAdapter.executeFunctions(props.ui.onClose);
}

const outerColor = computed<string>(() => {
	let cursor = '';
	if (props.ui.action.v != 0) {
		cursor = 'cursor-pointer ';
	}

	switch (props.ui.color.v) {
		case 'red':
			return cursor + 'text-red-800 bg-red-100';
		case 'green':
			return cursor + 'text-green-800 bg-green-100';
		case 'yellow':
			return cursor + 'text-yellow-800 bg-yellow-100';
		default:
			return cursor + 'text-gray-800 bg-gray-100';
	}
});

const innerColor = computed<string>(() => {
	switch (props.ui.color.v) {
		case 'red':
			return 'hover:bg-red-200 hover:text-red-900 text-red-400';
		case 'green':
			return 'hover:bg-green-200 hover:text-green-900';
		case 'yellow':
			return 'hover:bg-yellow-200 hover:text-yellow-900';
		default:
			return 'hover:bg-gray-200 hover:text-gray-900';
	}
});
</script>

<template>
	<span
		v-if="ui.visible.v"
		@click="onClick"
		:class="outerColor"
		class="me-2 inline-flex items-center rounded px-2 py-1 text-sm font-medium"
	>
		{{ props.ui.caption.v }}
		<button
			v-if="props.ui.onClose.v"
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
