<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm dark:text-white">{{ props.ui.label.value }}</span>

		<div
			class="slider"
			:class="{'slider-disabled': props.ui.disabled.value}"
		>
			<div class="relative flex items-center h-4">
				<!-- Slider track -->
				<div ref="sliderTrack" class="slider-track w-full"></div>
				<!-- Left slider thumb -->
				<div
					class="slider-thumb slider-thumb-start absolute left-0 size-4 rounded-full bg-ora-orange z-0"
					:class="{
						'slider-thumb-dragging': startDragging,
						'slider-thumb-disabled': !props.ui.startInitialized.value,
					}"
					:style="`--slider-thumb-start-offset: ${sliderThumbStartOffset}px;`"
					:tabindex="props.ui.disabled.value ? '-1' : '0'"
					@mousedown="startSliderThumbPressed"
					@touchstart="startSliderThumbPressed"
				></div>
				<!-- Right slider thumb -->
				<div
					class="slider-thumb slider-thumb-end absolute left-0 size-4 rounded-full bg-ora-orange z-10"
					:class="{
						'slider-thumb-dragging': endDragging,
						'slider-thumb-disabled': !props.ui.endInitialized.value,
					}"
					:style="`--slider-thumb-end-offset: ${sliderThumbEndOffset}px;`"
					:tabindex="props.ui.disabled.value ? '-1' : '0'"
					@mousedown="endSliderThumbPressed"
					@touchstart="endSliderThumbPressed"
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
import { onBeforeMount, onMounted, onUnmounted, ref } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveSlider;
}>();

const networkStore = useNetworkStore();
const sliderTrack = ref<HTMLElement|undefined>();
const startDragging = ref<boolean>(false);
const endDragging = ref<boolean>(false);
const scaleOffset = ref<number>(roundValue(props.ui.min.value));
const minRounded = ref<number>(0);
const maxRounded = ref<number>(roundValue(props.ui.max.value - scaleOffset.value));
const sliderStartValue = ref<number>(roundValue(props.ui.startValue.value - scaleOffset.value));
const sliderEndValue = ref<number>(roundValue(props.ui.endValue.value - scaleOffset.value));
const stepsizeRounded = ref<number>(roundValue(props.ui.stepsize.value));
const sliderThumbStartOffset = ref<number>(0);
const sliderThumbEndOffset = ref<number>(0);

onMounted(() => {
	initializeSliderThumbOffsets();

	addEventListeners();
});

onUnmounted(removeEventListeners);

function initializeSliderThumbOffsets(): void {
	sliderThumbStartOffset.value = sliderValueToOffset(sliderStartValue.value);
	sliderThumbEndOffset.value = sliderValueToOffset(sliderEndValue.value);
}

function addEventListeners(): void {
	document.addEventListener('mouseup', onMouseUp);
	document.addEventListener('touchend', onMouseUp);
	document.addEventListener('touchcancel', onMouseUp);
	document.addEventListener('touchmove', onTouchMove);
	document.addEventListener('mousemove', onMouseMove);
	window.addEventListener('resize', initializeSliderThumbOffsets, { passive: true });
}

function removeEventListeners(): void {
	document.removeEventListener('mouseup', onMouseUp);
	document.removeEventListener('touchend', onMouseUp);
	document.removeEventListener('touchcancel', onMouseUp);
	document.removeEventListener('touchmove', onTouchMove);
	document.removeEventListener('mousemove', onMouseMove);
	window.removeEventListener('resize', initializeSliderThumbOffsets);
}

function onMouseUp(): void {
	if (startDragging.value || endDragging.value) {
		startDragging.value = false;
		endDragging.value = false;
		submitSliderValues();
	}
}

function onMouseMove(event: MouseEvent): void {
	if (!sliderTrack.value || !startDragging.value && !endDragging.value) {
		return;
	}
	handleSliderThumbDrag(event.x, sliderTrack.value.getBoundingClientRect().x, sliderTrack.value.offsetWidth);
}

function onTouchMove(event: TouchEvent): void {
	const touchLocation = event.touches.item(0);
	if (!touchLocation || !sliderTrack.value || !startDragging.value && !endDragging.value) {
		return;
	}
	handleSliderThumbDrag(touchLocation.clientX, sliderTrack.value.getBoundingClientRect().x, sliderTrack.value.offsetWidth);
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
	const continuousValue = sliderOffsetPercentage * maxRounded.value;
	return getDiscreteValue(continuousValue);
}

function getDiscreteValue(continuousValue: number): number {
	let validValueBelow: number = minRounded.value;
	for (let validValue = minRounded.value; validValue <= continuousValue; validValue += stepsizeRounded.value) {
		validValueBelow = roundValue(validValue);
	}
	const validValueAbove = roundValue(validValueBelow + stepsizeRounded.value);
	if (validValueAbove > roundValue(props.ui.max.value - scaleOffset.value) || continuousValue - validValueBelow < validValueAbove - continuousValue) {
		return roundValue(validValueBelow);
	}
	return roundValue(validValueAbove);
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
				value: roundValue(sliderStartValue.value + scaleOffset.value),
			},
			{
				...props.ui.endValue,
				value: roundValue(sliderEndValue.value + scaleOffset.value),
			}
		], [props.ui.onChanged]);
}
</script>

<style scoped>
.slider.slider-disabled .slider-thumb {
	@apply bg-disabled-text;
}

.slider-track {
	@apply border-b border-b-black;
	@apply dark:border-b-white;
}

.slider.slider-disabled .slider-track {
	@apply border-b-disabled-background;
}

.slider:not(.slider-disabled) .slider-thumb.slider-thumb-disabled {
	@apply bg-black;
	@apply dark:bg-white;
}

.slider:not(.slider-disabled) .slider-thumb.slider-thumb-disabled:hover,
.slider:not(.slider-disabled) .slider-thumb.slider-thumb-disabled:focus-visible,
.slider:not(.slider-disabled) .slider-thumb.slider-thumb-disabled.slider-thumb-dragging {
	@apply bg-ora-orange;
}

.slider {
	@apply rounded-full p-2 -mx-2;
}

.slider-thumb {
	@apply select-none;
}

.slider-thumb:focus-visible:not(:hover) {
	@apply outline-none outline-black outline-offset-2 ring-white ring-2;
}

.slider:not(.slider-disabled) .slider-thumb:hover,
.slider:not(.slider-disabled) .slider-thumb.slider-thumb-dragging {
	@apply outline-none ring-8 ring-ora-orange ring-opacity-15;
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
