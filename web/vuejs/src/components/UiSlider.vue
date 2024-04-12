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
					class="slider-thumb slider-thumb-start absolute left-0 size-4 rounded-full bg-ora-orange z-0"
					:class="{'slider-thumb-dragging': startDragging}"
					:style="`--slider-thumb-start-offset: ${sliderThumbStartOffset}px;`"
					@mousedown="startSliderThumbPressed"
				></div>
				<!-- Right slider thumb -->
				<div
					class="slider-thumb slider-thumb-end absolute left-0 size-4 rounded-full bg-ora-orange z-10"
					:class="{'slider-thumb-dragging': endDragging}"
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
import { onBeforeMount, onMounted, ref } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveSlider;
}>();

const networkStore = useNetworkStore();
const sliderTrack = ref<HTMLElement|undefined>();
const sliderStartValue = ref<number>(0);
const sliderEndValue = ref<number>(0);
const startDragging = ref<boolean>(false);
const endDragging = ref<boolean>(false);
const minRounded = ref<number>(roundValue(props.ui.min.value));
const maxRounded = ref<number>(roundValue(props.ui.max.value));
const stepsizeRounded = ref<number>(roundValue(props.ui.stepsize.value));
const initialStartValueRounded = ref<number>(roundValue(props.ui.startValue.value));
const initialEndValueRounded = ref<number>(roundValue(props.ui.endValue.value));
const sliderThumbStartOffset = ref<number>(0);
const sliderThumbEndOffset = ref<number>(0);

onBeforeMount(() => {
	if (!props.ui.initialized.value) {
		sliderStartValue.value = minRounded.value;
		return;
	}
	sliderStartValue.value = calculateSliderValue(initialStartValueRounded.value);
	sliderEndValue.value = calculateSliderValue(initialEndValueRounded.value);
});

function calculateSliderValue(initialValue: number): number {
	// Limit initial value to min and max value
	const bounded = Math.max(Math.min(initialValue, maxRounded.value), minRounded.value);
	// Calculate valid initial value based on the step size and minimum value (always rounding down to the next valid value)
	const validated = bounded - (bounded - minRounded.value) % stepsizeRounded.value;
	// Get rid of rounding errors by using 2 decimal places at most
	return roundValue(validated);
}

onMounted(() => {
	sliderThumbStartOffset.value = sliderValueToOffset(sliderStartValue.value);
	sliderThumbEndOffset.value = sliderValueToOffset(sliderEndValue.value);

	addEventListeners();
});

function addEventListeners(): void {
	document.addEventListener('mouseup', () => {
		startDragging.value = false;
		endDragging.value = false;
		submitSliderValues();
	});
	document.addEventListener('mousemove', (event: MouseEvent) => {
		if (!sliderTrack.value || !startDragging.value && !endDragging.value) {
			return;
		}
		handleSliderThumbDrag(event.x, sliderTrack.value.getBoundingClientRect().x, sliderTrack.value.offsetWidth);
	});
}

/**
 * Maps a slider value to a pixel offset value for the corresponding slider thumb
 *
 * @param sliderValue The slider value to map
 */
function sliderValueToOffset(sliderValue: number): number {
	if (!sliderTrack.value) {
		return 0;
	}
	const sliderValuePercentage = sliderValue / maxRounded.value;
	return sliderTrack.value.offsetWidth * sliderValuePercentage;
}

/**
 * Maps a pixel offset value of a slider thumb to its corresponding slider value
 *
 * @param sliderThumbOffset The pixel offset value of a slider thumb to map
 */
function offsetToSliderValue(sliderThumbOffset: number): number {
	if (!sliderTrack.value) {
		return 0;
	}
	const sliderOffsetPercentage = sliderThumbOffset / sliderTrack.value.offsetWidth;
	const continuousValue = maxRounded.value * sliderOffsetPercentage;
	return getDiscreteValue(continuousValue);
}

function getDiscreteValue(continuousValue: number): number {
	let validValueBelow: number = minRounded.value;
	for (let validValue = minRounded.value; validValue <= continuousValue; validValue += stepsizeRounded.value) {
		validValueBelow = validValue;
	}
	const validValueAbove = validValueBelow + stepsizeRounded.value;
	if (continuousValue - validValueBelow < validValueAbove - continuousValue) {
		return validValueBelow;
	}
	return validValueAbove;
}

function startSliderThumbPressed(): void {
	if (!props.ui.disabled.value) {
		startDragging.value = true;
	}
}

function endSliderThumbPressed(): void {
	if (props.ui.disabled.value || !sliderTrack.value) {
		return;
	}
	if (sliderThumbStartOffset.value === sliderTrack.value.offsetWidth) {
		// Drag start slider thumb because of higher z-index of end slider thumb
		startDragging.value = true;
	} else {
		endDragging.value = true;
	}
}

function handleSliderThumbDrag(mouseX: number, sliderTrackOffsetX: number, sliderTrackOffsetWidth: number): void {
	if (startDragging.value) {
		const continuousOffset = Math.max(0, Math.min(sliderThumbEndOffset.value, mouseX - sliderTrackOffsetX));
		sliderStartValue.value = offsetToSliderValue(continuousOffset);
		sliderThumbStartOffset.value = sliderValueToOffset(sliderStartValue.value);
	} else if (endDragging.value) {
		const continuousOffset = Math.max(sliderThumbStartOffset.value, Math.min(mouseX - sliderTrackOffsetX, sliderTrackOffsetWidth));
		sliderEndValue.value = offsetToSliderValue(continuousOffset);
		sliderThumbEndOffset.value = sliderValueToOffset(sliderEndValue.value);
	}
}

function roundValue(value: number): number {
	return Math.round(value * 100) / 100;
}

function submitSliderValues(): void {
	networkStore.invokeFunctionsAndSetProperties([
			{
				...props.ui.startValue,
				value: roundValue(sliderStartValue.value),
			},
			{
				...props.ui.endValue,
				value: roundValue(sliderEndValue.value),
			}
		], [props.ui.onChanged]);
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

.slider-thumb {
	@apply select-none;
}

.slider:not(.slider-disabled) .slider-thumb:hover,
.slider:not(.slider-disabled) .slider-thumb.slider-thumb-dragging {
	@apply ring-8 ring-ora-orange ring-opacity-15;
	@apply dark:ring-opacity-25;
}

.slider:not(.slider-disabled) .slider-thumb.slider-thumb-dragging {
	@apply ring-opacity-25;
	@apply dark:ring-opacity-35 !important;
}

.slider-thumb-start {
	left: calc(var(--slider-thumb-start-offset) - 0.5rem);
}

.slider-thumb-end {
	left: calc(var(--slider-thumb-end-offset) - 0.5rem);
}
</style>
