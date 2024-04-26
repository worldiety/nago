<script lang="ts" setup>
import type {TextArea} from "@/shared/protocol/gen/textArea";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: TextArea;
}>();

const serviceAdapter = useServiceAdapter();

function valueChanged(event: any) {
	props.ui.value.v = event.target.value;
	serviceAdapter.setPropertiesAndCallFunctions([props.ui.value], [props.ui.onTextChanged]);
}

function isErr(): boolean {
	return props.ui.error.v != '';
}
</script>

<template>
	<div>
		<label
			:for="props.ui.id.toString()"
			:class="isErr() ? 'text-red-700 dark:text-red-500' : 'text-gray-900 dark:text-white'"
			class="mb-2 block text-sm font-medium"
			>{{ props.ui.label.v }}</label
		>
		<textarea
			:disabled="props.ui.disabled.v"
			@input="valueChanged"
			:value="props.ui.value.v"
			type="text"
			:id="props.ui.id.toString()"
			:rows="props.ui.rows.v"
			:class="
				isErr()
					? 'block w-full rounded-lg border border-red-500 bg-red-50 p-2.5 text-sm text-red-900 placeholder-red-700 focus:border-red-500 focus:ring-red-500 dark:border-red-500 dark:bg-gray-700 dark:text-red-500 dark:placeholder-red-500'
					: 'block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500'
			"
		>
		</textarea>
		<p v-if="isErr()" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.v }}</p>
		<p v-if="!isErr()" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.v }}</p>
	</div>
</template>
