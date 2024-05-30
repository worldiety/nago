<script lang="ts" setup>
import type {TextArea} from "@/shared/protocol/ora/textArea";
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
	<div v-if="ui.visible.v">
		<label
			:for="props.ui.id.toString()"
			:class="isErr() ? 'text-red-700' : 'text-gray-900'"
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
			class="rounded-lg bg-transparent border text-sm p-2.5 w-full"
			:class="
				isErr()
					? 'border-error'
					: 'border-gray-300'
			"
		>
		</textarea>
		<p v-if="isErr()" class="mt-2 text-sm text-red-600">{{ props.ui.error.v }}</p>
		<p v-if="!isErr()" class="mt-2 text-sm text-gray-500">{{ props.ui.hint.v }}</p>
	</div>
</template>
