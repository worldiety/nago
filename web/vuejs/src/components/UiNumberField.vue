<script setup lang="ts">
import type { LiveNumberField } from '@/shared/model/liveNumberField';
import { ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { PropertyInt } from '@/shared/model/propertyInt';
import InputWrapper from '@/components/shared/InputWrapper.vue';

const props = defineProps<{
	ui: LiveNumberField;
}>();

const networkStore = useNetworkStore();
const inputValue = ref<string>(props.ui.value.value.toString(10));

/**
 * Validates the input value and submits it, if it is valid.
 * The '-' sign and the empty string are treated as 0.
 * If the input value is invalid, the value gets reset to the last known valid value.
 */
watch(inputValue, (newValue, oldValue) => {
	if (newValue === '' || newValue == '-') {
		newValue = '0';
	} else if (!newValue.match(/^-?[0-9]+$/)) {
		inputValue.value = oldValue;
		return;
	}
	const updatedValueProperty: PropertyInt = {
		...props.ui.value,
		value: parseInt(newValue, 10),
	};
	networkStore.invokeFuncAndSetProp(updatedValueProperty, props.ui.onValueChanged);
});
</script>

<template>
	<div>
		<InputWrapper
			:simple="props.ui.simple.value"
			:label="props.ui.label.value"
			:error="props.ui.error.value"
			:hint="props.ui.hint.value"
			:disabled="props.ui.disabled.value"
		>
			<input
				v-model="inputValue"
				type="text"
				class="input-field"
				inputmode="numeric"
				:placeholder="props.ui.placeholder.value"
				:disabled="props.ui.disabled.value"
			>
		</InputWrapper>
	</div>
</template>
