<script setup lang="ts">
import { ref, watch } from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';
import type {NumberField} from "@/shared/protocol/ora/numberField";
import type {Property} from "@/shared/protocol/ora/property";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: NumberField
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.value.v);
let timeout: number|null = null;

watch(() => props.ui.value.v, (newValue) => {
	inputValue.value = newValue;
});

/**
 * Validates the input value and submits it, if it is valid.
 * The '-' sign and the empty string are treated as 0.
 * If the input value is invalid, the value gets reset to the last known valid value.
 */
watch(inputValue, (newValue, oldValue) => {
	if (newValue === '' || newValue == '-') {
		inputValue.value = '0';
	} else if (!newValue.match(/^-?[0-9]+$/)) {
		inputValue.value = oldValue;
		return;
	}

	if (timeout !== null) {
		return;
	}
	timeout = window.setTimeout(() => {
		const updatedValueProperty: Property<string> = {
			...props.ui.value,
			v: inputValue.value,
		};
		serviceAdapter.setPropertiesAndCallFunctions([updatedValueProperty], [props.ui.onValueChanged]);
		timeout = null;
	}, 500);
});
</script>

<template>
	<div>
		<InputWrapper
			:simple="props.ui.simple.v"
			:label="props.ui.label.v"
			:error="props.ui.error.v"
			:hint="props.ui.hint.v"
			:disabled="props.ui.disabled.v"
		>
			<div class="relative">
				<input
					v-model="inputValue"
					type="text"
					class="input-field"
					inputmode="numeric"
					:placeholder="props.ui.placeholder.v"
					:disabled="props.ui.disabled.v"
				>
				<div v-if="inputValue" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="0" @click="inputValue = ''" @keydown.enter="inputValue = ''" />
				</div>
			</div>
	</InputWrapper>

	</div>
</template>
