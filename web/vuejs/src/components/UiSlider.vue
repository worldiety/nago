<template>
	<div>
		<span v-if="props.ui.label.v" class="block mb-2 text-sm dark:text-white">{{ props.ui.label.v }}</span>

		<div
			class="slider"
			:class="{
				'slider-disabled': props.ui.disabled.v,
				'mb-6': props.ui.showLabel.v,
			}"
			:style="`--slider-thumb-start-offset: ${sliderThumbStartOffset}px; --slider-thumb-end-offset: ${sliderThumbEndOffset}px;`"
		>
			<div class="relative flex items-center h-4">
				<!-- Slider track -->
				<div ref="sliderTrack" class="slider-track w-full">
					<!-- Slider tick marks -->
					<template v-if="props.ui.showTickMarks.v">
						<div
							v-for="(sliderTickMark, index) in sliderTickMarks"
							:key="index"
							class="slider-tick-mark"
							:class="{'slider-tick-mark-in-range': sliderTickMark.withinRange}"
							:style="`--slider-tick-mark-offset: ${sliderTickMark.offset}px;`"
						></div>
					</template>
				</div>
				<!-- Left slider thumb -->
				<div
					v-if="props.ui.rangeMode.v"
					class="slider-thumb slider-thumb-start absolute left-0 size-4 rounded-full bg-ora-orange"
					:class="{
						'slider-thumb-dragging': startDragging,
						'slider-thumb-uninitialized': !props.ui.startInitialized.v,
						'z-10': !startDragging,
						'z-20': startDragging,
					}"
					:tabindex="props.ui.disabled.v ? '-1' : '0'"
					@mousedown="startSliderThumbPressed"
					@touchstart="startSliderThumbPressed"
					@keydown.left="decreaseStartSliderValue"
					@keydown.right="increaseStartSliderValue"
				>
					<div v-if="props.ui.showLabel.v && props.ui.startInitialized.v" class="slider-thumb-label">
						<span>{{ getSliderLabel(sliderStartValue + scaleOffset) }}</span>
					</div>
				</div>
				<!-- Slider thumb connector -->
				<div
					v-if="sliderThumbConnectorVisible"
					class="slider-thumb-connector absolute top-1/2 border-b border-b-ora-orange z-0"
				></div>
				<!-- Right slider thumb -->
				<div
					class="slider-thumb slider-thumb-end absolute left-0 size-4 rounded-full bg-ora-orange"
					:class="{
						'slider-thumb-dragging': endDragging,
						'slider-thumb-uninitialized': !props.ui.endInitialized.v,
						'z-10': !endDragging,
						'z-20': endDragging,
					}"
					:tabindex="props.ui.disabled.v ? '-1' : '0'"
					@mousedown="endSliderThumbPressed"
					@touchstart="endSliderThumbPressed"
					@keydown.left="decreaseEndSliderValue"
					@keydown.right="increaseEndSliderValue"
				>
					<div v-if="props.ui.showLabel.v && props.ui.endInitialized.v" class="slider-thumb-label" :class="endDragging ? 'z-10' : 'z-0'">
						<span>{{ getSliderLabel(sliderEndValue + scaleOffset) }}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.v" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.v }}</p>
		<p v-else-if="props.ui.hint.v" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.v }}</p>
	</div>
</template>

<script setup lang="ts">
import { computed, onBeforeMount, onMounted, onUnmounted, ref, watch } from 'vue';
import type { Slider } from "@/shared/protocol/ora/slider";
import { useServiceAdapter } from '@/composables/serviceAdapter';

interface SliderTickMark {
	offset: number;
	withinRange: boolean;
}

const props = defineProps<{
	ui: Slider;
}>();

const serviceAdapter = useServiceAdapter();
const sliderTrack = ref<HTMLElement|undefined>();
const startDragging = ref<boolean>(false);
const endDragging = ref<boolean>(false);
const scaleOffset = ref<number>(roundValue(props.ui.min.v));
const minRounded = ref<number>(0);
const maxRounded = ref<number>(roundValue(props.ui.max.v - scaleOffset.value));
const sliderStartValue = ref<number>(0);
const sliderEndValue = ref<number>(0);
const stepsizeRounded = ref<number>(roundValue(props.ui.stepsize.v));
const sliderThumbStartOffset = ref<number>(0);
const sliderThumbEndOffset = ref<number>(0);
const sliderTickMarks = ref<SliderTickMark[]>([]);

onBeforeMount(() => {
	initializeBoundaries();
})

onMounted(() => {
	initializeSliderThumbOffsets();

	addEventListeners();
});

onUnmounted(removeEventListeners);

watch(() => props.ui.min.v, (newValue) => {
	scaleOffset.value = roundValue(newValue);
	initializeBoundaries();
	initializeSliderThumbOffsets();
});

watch(() => props.ui.max.v, (newValue) => {
	maxRounded.value = roundValue(newValue - scaleOffset.value)
	initializeBoundaries();
	initializeSliderThumbOffsets();
});

watch(() => props.ui.stepsize.v, (newValue) => {
	stepsizeRounded.value = roundValue(newValue);
	initializeBoundaries();
	initializeSliderThumbOffsets();
});

watch(sliderThumbStartOffset, initializeSliderTickMarks);

watch(sliderThumbEndOffset, initializeSliderTickMarks);

watch(() => props.ui.endInitialized.v, initializeSliderTickMarks);

watch(() => props.ui.startInitialized.v, initializeSliderTickMarks);

const sliderThumbConnectorVisible = computed((): boolean => {
	return props.ui.showLabel.v && (props.ui.startInitialized.v && props.ui.endInitialized.v || !props.ui.rangeMode.v && props.ui.endInitialized.v);
});

function getSliderLabel(sliderValue: number): string {
	return sliderValue.toLocaleString(undefined, {
		minimumFractionDigits: 2,
	}) + props.ui.labelSuffix.v;
}

function initializeBoundaries(): void {
	const startValue = props.ui.rangeMode.v ? roundValue(props.ui.startValue.v - scaleOffset.value) : roundValue(minRounded.value);
	sliderStartValue.value = getDiscreteValue(startValue);
	const endValue = roundValue(props.ui.endValue.v - scaleOffset.value);
	sliderEndValue.value = getDiscreteValue(endValue);
}

function initializeSliderThumbOffsets(): void {
	sliderThumbStartOffset.value = sliderValueToOffset(sliderStartValue.value);
	sliderThumbEndOffset.value = sliderValueToOffset(sliderEndValue.value);
}

function initializeSliderTickMarks(): void {
	const updatedSliderTickMarks: SliderTickMark[] = [];
	const totalSteps = Math.floor(((maxRounded.value - minRounded.value) / stepsizeRounded.value) + 1);
	for (let i = 0; i < totalSteps; i++) {
		const tickMarkOffset = sliderValueToOffset(i * stepsizeRounded.value);
		const withinRange = props.ui.showLabel.v && props.ui.startInitialized.v && props.ui.endInitialized.v && sliderThumbStartOffset.value <= tickMarkOffset && tickMarkOffset <= sliderThumbEndOffset.value;
		updatedSliderTickMarks.push({
			offset: tickMarkOffset,
			withinRange: withinRange,
		});
	}
	sliderTickMarks.value = updatedSliderTickMarks;
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
	if (validValueAbove > roundValue(props.ui.max.v - scaleOffset.value) || continuousValue - validValueBelow < validValueAbove - continuousValue) {
		return validValueBelow;
	}
	return validValueAbove;
}

function startSliderThumbPressed(): void {
	if (!props.ui.disabled.v) {
		startDragging.value = true;
	}
}

function endSliderThumbPressed(): void {
	if (props.ui.disabled.v || !sliderTrack.value) {
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
	serviceAdapter.setPropertiesAndCallFunctions([
		{
			...props.ui.startValue,
			v: roundValue(sliderStartValue.value + scaleOffset.value),
		},
		{
			...props.ui.endValue,
			v: roundValue(sliderEndValue.value + scaleOffset.value),
		}
	], [props.ui.onChanged]);
}

function decreaseStartSliderValue(): void {
	sliderStartValue.value = getDiscreteValue(sliderStartValue.value - stepsizeRounded.value);
	sliderThumbStartOffset.value = sliderValueToOffset(sliderStartValue.value);
	submitSliderValues();
}

function increaseStartSliderValue(): void {
	sliderStartValue.value = Math.min(
		getDiscreteValue(sliderStartValue.value + stepsizeRounded.value),
		sliderEndValue.value,
	);
	sliderThumbStartOffset.value = sliderValueToOffset(sliderStartValue.value);
	submitSliderValues();
}

function decreaseEndSliderValue(): void {
	sliderEndValue.value = Math.max(
		getDiscreteValue(sliderEndValue.value - stepsizeRounded.value),
		sliderStartValue.value,
	);
	sliderThumbEndOffset.value = sliderValueToOffset(sliderEndValue.value);
	submitSliderValues();
}

function increaseEndSliderValue(): void {
	sliderEndValue.value = getDiscreteValue(sliderEndValue.value + stepsizeRounded.value);
	sliderThumbEndOffset.value = sliderValueToOffset(sliderEndValue.value);
	submitSliderValues();
}
</script>

<style scoped>
.slider.slider-disabled .slider-thumb {
	@apply bg-disabled-text;
}

.slider-track {
	@apply relative border-b border-b-black;
	@apply dark:border-b-white;
}

.slider.slider-disabled .slider-track {
	@apply border-b-disabled-background;
}

.slider-tick-mark {
	@apply absolute -top-[4px] border-l border-l-black h-[9px];
	@apply dark:border-l-white;
	left: var(--slider-tick-mark-offset);
}

.slider.slider-disabled .slider-tick-mark {
	@apply border-l-disabled-background;
}

.slider:not(.slider-disabled) .slider-tick-mark.slider-tick-mark-in-range {
	@apply border-l-ora-orange;
}

.slider.slider-disabled .slider-tick-mark.slider-tick-mark-in-range {
	@apply border-l-disabled-text;
}

.slider:not(.slider-disabled) .slider-thumb.slider-thumb-uninitialized {
	@apply bg-black;
	@apply dark:bg-white;
}

.slider:not(.slider-disabled) .slider-thumb.slider-thumb-uninitialized:hover,
.slider:not(.slider-disabled) .slider-thumb.slider-thumb-uninitialized.slider-thumb-dragging {
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

.slider .slider-thumb-connector {
	width: calc(var(--slider-thumb-end-offset) - var(--slider-thumb-start-offset));
	left: calc(var(--slider-thumb-start-offset));
}

.slider.slider-disabled .slider-thumb-connector {
	@apply border-b-disabled-text;
}

.slider-thumb-label {
	@apply absolute left-0 right-0 flex justify-center text-ora-orange text-sm whitespace-nowrap overflow-visible;
	top: 150%;
}

.slider.slider-disabled .slider-thumb-label {
	@apply text-disabled-text;
}

.slider-thumb-label > span {
	@apply bg-white rounded-lg px-1;
	@apply dark:bg-darkmode-gray;
}
</style>
