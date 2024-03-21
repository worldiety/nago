<script setup lang="ts">
import type { LiveNumberField } from '@/shared/model/liveNumberField';
import { ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { PropertyInt } from '@/shared/model/propertyInt';

const props = defineProps<{
	ui: LiveNumberField;
}>();

const networkStore = useNetworkStore();
const inputValue = ref<string>('');

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
		<label
			:for="props.ui.id.toString()"
			class="block mb-2 text-sm font-medium"
		>
			{{ props.ui.label.value }}
		</label>

		<input v-model="inputValue" type="text" class="input-field w-full" inputmode="numeric" :disabled="props.ui.disabled.value">

		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>
