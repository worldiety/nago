<script setup lang="ts">
import type { LiveSlider } from '@/shared/model/liveSlider';
import { ref } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveSlider;
}>();

const networkStore = useNetworkStore();
const sliderValue = ref<number>(props.ui.value.value);

function submitSliderValue(): void {
	networkStore.invokeSetProp({
		...props.ui.value,
		value: sliderValue.value,
	});
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</span>

		<input
			v-model="sliderValue"
			type="range"
			:min="props.ui.min.value"
			:max="props.ui.max.value"
			:step="props.ui.stepsize.value"
			:disabled="props.ui.disabled.value"
			@mouseup="submitSliderValue"
			@touchend="submitSliderValue"
			@keyup.left.right="submitSliderValue"
		/>

		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>
