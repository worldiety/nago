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

onBeforeMount(() => {
	if (!props.ui.initialized.value) {
		sliderValue.value = props.ui.min.value;
		return;
	}
	// Limit initial value to min and max value
	const bounded = Math.max(Math.min(props.ui.value.value, props.ui.max.value), props.ui.min.value);
	// Calculate valid initial value based on the step size and minimum value (always rounding down)
	sliderValue.value = bounded - (bounded - props.ui.min.value) % props.ui.stepsize.value;
});

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
			:min="props.ui.min.value"
			:max="props.ui.max.value"
			:step="props.ui.stepsize.value"
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
