<script setup lang="ts">
import type { LiveSlider } from '@/shared/model/liveSlider';
import { onBeforeMount, ref } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveSlider;
}>();

const networkStore = useNetworkStore();
const sliderValue = ref<number>(0);
const dragging = ref<boolean>(false);
const minRounded = ref<number>(roundValue(props.ui.min.value));
const maxRounded = ref<number>(roundValue(props.ui.max.value));
const stepsizeRounded = ref<number>(roundValue(props.ui.stepsize.value));
const initialValueRounded = ref<number>(roundValue(props.ui.value.value));

onBeforeMount(() => {
	if (!props.ui.initialized.value) {
		sliderValue.value = minRounded.value;
		return;
	}
	// Limit initial value to min and max value
	const bounded = Math.max(Math.min(initialValueRounded.value, maxRounded.value), minRounded.value);
	// Calculate valid initial value based on the step size and minimum value (always rounding down to the next valid value)
	const validated = bounded - (bounded - minRounded.value) % stepsizeRounded.value;
	// Get rid of rounding errors
	sliderValue.value = roundValue(validated);
});

function roundValue(value: number): number {
	return Math.round(value * 100) / 100;
}

function submitSliderValue(): void {
	dragging.value = false;
	networkStore.invokeFuncAndSetProp({
		...props.ui.value,
		value: sliderValue.value,
	}, props.ui.onChanged);
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm">{{ props.ui.label.value }}</span>

		<input
			v-model="sliderValue"
			type="range"
			:min="minRounded"
			:max="maxRounded"
			:step="stepsizeRounded"
			:disabled="props.ui.disabled.value"
			:class="{'slider-dragging': dragging, 'slider-uninitialized': !props.ui.initialized.value}"
			class="px-2 -ml-2"
			@mousedown="dragging = true"
			@touchstart="dragging = true"
			@mouseup="submitSliderValue"
			@touchend="submitSliderValue"
			@keyup.left.right="submitSliderValue"
		/>

		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>
