<script lang="ts" setup>
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveTextArea } from '@/shared/model/liveTextArea';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveTextArea;
	page: LivePage;
}>();

const networkStore = useNetworkStore();

function valueChanged(event: any) {
	props.ui.value.value = event.target.value;
	networkStore.invokeFuncAndSetProp(props.ui.value, props.ui.onTextChanged);
}

function isErr(): boolean {
	return props.ui.error.value != '';
}
</script>

<template>
	<div>
		<label
			:for="props.ui.id.toString()"
			:class="isErr() ? 'text-red-700 dark:text-red-500' : 'text-gray-900 dark:text-white'"
			class="mb-2 block text-sm font-medium"
			>{{ props.ui.label.value }}</label
		>
		<textarea
			:disabled="props.ui.disabled.value"
			@input="valueChanged"
			:value="props.ui.value.value"
			type="text"
			:id="props.ui.id.toString()"
			:rows="props.ui.rows.value"
			:class="
				isErr()
					? 'block w-full rounded-lg border border-red-500 bg-red-50 p-2.5 text-sm text-red-900 placeholder-red-700 focus:border-red-500 focus:ring-red-500 dark:border-red-500 dark:bg-gray-700 dark:text-red-500 dark:placeholder-red-500'
					: 'block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500'
			"
		>
		</textarea>
		<p v-if="isErr()" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-if="!isErr()" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>
