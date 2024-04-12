<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm">{{ props.ui.label.value }}</span>

		<div
			class="slider"
			:class="{'slider-disabled': props.ui.disabled.value}"
			:tabindex="props.ui.disabled.value ? '-1' : '0'"
		>
			<div class="relative flex items-center h-4">
				<!-- Slider track -->
				<div ref="sliderTrack" class="slider-track border-b border-b-black w-full"></div>
				<!-- Left slider thumb -->
				<div
					class="slider-thumb slider-thumb-start absolute left-0 size-4 rounded-full bg-ora-orange"
					:style="`--slider-thumb-start-offset: ${sliderThumbStartOffset}px;`"
					@mousedown="startSliderThumbPressed"
				></div>
				<!-- Right slider thumb -->
				<div
					class="slider-thumb slider-thumb-end absolute right-0 size-4 rounded-full bg-ora-orange"
					:style="`--slider-thumb-end-offset: ${sliderThumbEndOffset}px;`"
					@mousedown="endSliderThumbPressed"
				></div>
			</div>
		</div>

		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>

<script setup lang="ts">
import type { LiveSlider } from '@/shared/model/liveSlider';
import { onBeforeMount, onMounted, ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveSlider;
}>();

const networkStore = useNetworkStore();
const sliderTrack = ref<HTMLElement|undefined>();
const sliderValue = ref<number>(0);
const startDragging = ref<boolean>(false);
const endDragging = ref<boolean>(false);
const minRounded = ref<number>(roundValue(props.ui.min.value));
const maxRounded = ref<number>(roundValue(props.ui.max.value));
const stepsizeRounded = ref<number>(roundValue(props.ui.stepsize.value));
const initialValueRounded = ref<number>(roundValue(props.ui.value.value));
const sliderThumbStartOffset = ref<number>(0);
const sliderThumbEndOffset = ref<number>(0);

onBeforeMount(() => {
	if (!props.ui.initialized.value) {
		sliderValue.value = minRounded.value;
		return;
	}
	// Limit initial value to min and max value
	const bounded = Math.max(Math.min(initialValueRounded.value, maxRounded.value), minRounded.value);
	// Calculate valid initial value based on the step size and minimum value (always rounding down to the next valid value)
	const validated = bounded - (bounded - minRounded.value) % stepsizeRounded.value;
	// Get rid of rounding errors by using 2 decimal places at most
	sliderValue.value = roundValue(validated);
});

onMounted(() => {
	document.addEventListener('mouseup', () => {
		startDragging.value = false;
		endDragging.value = false;
	});
	document.addEventListener('mousemove', (event: MouseEvent) => {
		if (!sliderTrack.value || !startDragging.value && !endDragging.value) {
			return;
		}
		handleSliderThumbDrag(event.x, sliderTrack.value.getBoundingClientRect().x, sliderTrack.value.offsetWidth);
	});
});

function startSliderThumbPressed(): void {
	if (!props.ui.disabled.value) {
		startDragging.value = true;
	}
}

function endSliderThumbPressed(): void {
	if (!props.ui.disabled.value) {
		endDragging.value = true;
	}
}

function handleSliderThumbDrag(mouseX: number, sliderTrackOffsetX: number, sliderTrackOffsetWidth: number): void {
	// TODO: Limit offsets to keep start always before and and offsets between lower and upper bound of slider track
	if (startDragging.value) {
		sliderThumbStartOffset.value = mouseX - sliderTrackOffsetX;
	} else if (endDragging.value) {
		sliderThumbEndOffset.value = (mouseX - sliderTrackOffsetX - sliderTrackOffsetWidth) * -1;
	}
}

function roundValue(value: number): number {
	return Math.round(value * 100) / 100;
}

function submitSliderValue(): void {
	startDragging.value = false;
	networkStore.invokeFunctionsAndSetProperties([{
		...props.ui.value,
		value: sliderValue.value,
	}], [props.ui.onChanged]);
}
</script>

<style scoped>
.slider.slider-disabled .slider-thumb {
	@apply bg-disabled-text;
}

.slider.slider-disabled .slider-track {
	@apply border-b-disabled-background
}

.slider {
	@apply rounded-full p-2 -mx-2;
}

.slider:focus-visible {
	@apply outline-none outline-2 outline-offset-2 outline-black ring-white ring-2;
}

.slider:not(.slider-disabled) .slider-thumb:hover {
	@apply ring-8 ring-ora-orange ring-opacity-15;
	@apply dark:ring-opacity-25;
}

.slider:not(.slider-disabled) .slider-thumb:active {
	@apply ring-opacity-25;
	@apply dark:ring-opacity-35 !important;
}

.slider-thumb-start {
	left: calc(var(--slider-thumb-start-offset) - 0.5rem);
}

.slider-thumb-end {
	right: calc(var(--slider-thumb-end-offset) - 0.5rem);
}
</style>
