<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div :style="frameStyles">
		<InputWrapper
			:wrapper-style="InputWrapperStyle.BASIC"
			:label="ui.label"
			:error="ui.errorText"
			:help="ui.supportingText"
			:disabled="ui.disabled"
		>
			<div
				class="slider"
				:class="{ 'disabled': ui.disabled, 'error': ui.errorText, 'range-mode': isRangeMode }"
				:aria-valuemin="(ui.min ?? 0) + (ui.unit ?? '')"
				:aria-valuemax="(ui.max ?? 0) + (ui.unit ?? '')"
				:aria-valuenow="
					isRangeMode
						? `${$t('slider.from')} ${valueFrom}${ui.unit ?? ''} ${$t('slider.to')} ${valueTo}${ui.unit ?? ''}`
						: valueFrom + (ui.unit ?? '')
				"
				role="slider"
			>
				<div ref="bar" class="bar">
					<template v-if="ui.showMarkers">
						<div
							v-for="i in stepsBetween"
							:key="`slider_step_${i}`"
							class="step"
							:class="{ active: isStepActive(i) }"
							:style="`left: ${i * stepWidth * 100}%;`"
						></div>
					</template>

					<div class="bar-left" :style="barLeftStyles"></div>
					<div v-if="isRangeMode" class="bar-middle" :style="barMiddleStyles"></div>
					<div class="bar-right" :style="barRightStyles"></div>

					<!-- From grabber (single grabber in non-range mode, left grabber in range mode) -->
					<div
						class="grabber"
						:style="grabberFromStyles"
						:class="{ dragging: isDraggingFrom }"
						:tabindex="ui.disabled ? -1 : 0"
						@pointerdown.prevent="onPointerDownFrom"
						@keydown.left="prevStepFrom"
						@keydown.right="nextStepFrom"
					>
						<span ref="grabberDot" class="dot"></span>
						<span class="value">{{ valueFrom }}{{ ui.unit }}</span>
					</div>

					<!-- To grabber (only in range mode) -->
					<div
						v-if="isRangeMode"
						class="grabber"
						:style="grabberToStyles"
						:class="{ dragging: isDraggingTo }"
						:tabindex="ui.disabled ? -1 : 0"
						@pointerdown.prevent="onPointerDownTo"
						@keydown.left="prevStepTo"
						@keydown.right="nextStepTo"
					>
						<span class="dot"></span>
						<span class="value">{{ valueTo }}{{ ui.unit }}</span>
					</div>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref, onUnmounted, watch } from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { InputWrapperStyle } from '@/components/shared/inputWrapperStyle';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Slider, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Slider;
}>();

const serviceAdapter = useServiceAdapter();

const bar = ref<HTMLElement>();
const grabberDot = ref<HTMLElement>();

const step = ref(props.ui.step ?? 1);
const valueFrom = ref(props.ui.value?.from ?? 0);
const valueTo = ref(props.ui.value?.to ?? 0);

const isRangeMode = computed<boolean>(() => !!props.ui.rangeMode);

const isDraggingFrom = ref(false);
const isDraggingTo = ref(false);
let lastSnappedFrom: number | null = null;
let lastSnappedTo: number | null = null;

const stepsBetween = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return Math.floor((max - min) / step.value) - 1;
});

const stepWidth = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return step.value / (max - min);
});

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);
	return styles.join(';');
});

const valueFromRatio = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return (valueFrom.value - min) / (max - min);
});

const valueToRatio = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return (valueTo.value - min) / (max - min);
});

const grabberHalfWidth = computed<number>(() => {
	if (!grabberDot.value) return 0;
	return grabberDot.value.clientWidth / 2;
});

const grabberFromStyles = computed<string>(() => {
	return `left: ${valueFromRatio.value * 100}%;`;
});

const grabberToStyles = computed<string>(() => {
	return `left: ${valueToRatio.value * 100}%;`;
});

const barLeftStyles = computed<string>(() => {
	return `width: calc(${valueFromRatio.value * 100}% - ${grabberHalfWidth.value}px);`;
});

const barMiddleStyles = computed<string>(() => {
	return `left: calc(${valueFromRatio.value * 100}% + ${grabberHalfWidth.value}px); right: calc(${(1 - valueToRatio.value) * 100}% + ${grabberHalfWidth.value}px);`;
});

const barRightStyles = computed<string>(() => {
	const ratio = isRangeMode.value ? valueToRatio.value : valueFromRatio.value;
	return `width: calc(${(1 - ratio) * 100}% - ${grabberHalfWidth.value}px);`;
});

function isStepActive(i: number): boolean {
	const pos = (props.ui.min ?? 0) + step.value * i;
	if (isRangeMode.value) {
		return pos >= valueFrom.value && pos <= valueTo.value;
	}
	return valueFrom.value >= pos;
}

function submitValue() {
	if (isRangeMode.value) {
		const json = JSON.stringify({ From: valueFrom.value, To: valueTo.value });
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), json));
	} else {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), `${valueFrom.value}`)
		);
	}
}

function getDecimalPlacesByStepSize(): number {
	const split = `${step.value}`.split('.');
	if (split.length === 1) return 0;
	return split[1].length;
}

function fixDecimals(num: number): number {
	return parseFloat(num.toFixed(getDecimalPlacesByStepSize()));
}

function snapToStep(clientX: number): number {
	const rect = bar.value!.getBoundingClientRect();
	const ratio = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	const totalSteps = (max - min) / step.value;
	const snappedStep = Math.round(ratio * totalSteps);
	return fixDecimals(min + snappedStep * step.value);
}

// --- Left or only grabber ---
function prevStepFrom() {
	if (props.ui.disabled || valueFrom.value === props.ui.min) return;
	const startVal = valueFrom.value;
	valueFrom.value = Math.max(props.ui.min ?? 0, fixDecimals(startVal - step.value));
	submitValue();
}

function nextStepFrom() {
	if (props.ui.disabled) return;
	const ceiling = isRangeMode.value ? valueTo.value - step.value : (props.ui.max ?? 0);
	const startVal = valueFrom.value;
	valueFrom.value = Math.min(ceiling, fixDecimals(startVal + step.value));
	submitValue();
}

function onPointerMoveFrom(event: PointerEvent) {
	const snapped = snapToStep(event.clientX);
	const clamped = isRangeMode.value ? Math.min(snapped, fixDecimals(valueTo.value - step.value)) : snapped;
	if (clamped !== lastSnappedFrom) {
		lastSnappedFrom = clamped;
		valueFrom.value = clamped;
	}
}

function onPointerUpFrom() {
	if (lastSnappedFrom !== null && lastSnappedFrom !== (props.ui.value?.from ?? 0)) {
		submitValue();
	}
	isDraggingFrom.value = false;
	document.removeEventListener('pointermove', onPointerMoveFrom);
	document.removeEventListener('pointerup', onPointerUpFrom);
}

function onPointerDownFrom() {
	if (props.ui.disabled) return;
	isDraggingFrom.value = true;
	lastSnappedFrom = valueFrom.value;
	document.addEventListener('pointermove', onPointerMoveFrom);
	document.addEventListener('pointerup', onPointerUpFrom);
}

// --- Right grabber (range mode only) ---
function prevStepTo() {
	if (props.ui.disabled) return;
	const startVal = valueTo.value;
	const floor = valueFrom.value + step.value;
	valueTo.value = Math.max(floor, fixDecimals(startVal - step.value));
	submitValue();
}

function nextStepTo() {
	if (props.ui.disabled || valueTo.value === props.ui.max) return;
	const startVal = valueTo.value;
	valueTo.value = Math.min(props.ui.max ?? 0, fixDecimals(startVal + step.value));
	submitValue();
}

function onPointerMoveTo(event: PointerEvent) {
	const snapped = snapToStep(event.clientX);
	const clamped = Math.max(snapped, fixDecimals(valueFrom.value + step.value));
	if (clamped !== lastSnappedTo) {
		lastSnappedTo = clamped;
		valueTo.value = clamped;
	}
}

function onPointerUpTo() {
	if (lastSnappedTo !== null && lastSnappedTo !== (props.ui.value?.to ?? 0)) {
		submitValue();
	}
	isDraggingTo.value = false;
	document.removeEventListener('pointermove', onPointerMoveTo);
	document.removeEventListener('pointerup', onPointerUpTo);
}

function onPointerDownTo() {
	if (props.ui.disabled) return;
	isDraggingTo.value = true;
	lastSnappedTo = valueTo.value;
	document.addEventListener('pointermove', onPointerMoveTo);
	document.addEventListener('pointerup', onPointerUpTo);
}

onUnmounted(() => {
	document.removeEventListener('pointermove', onPointerMoveFrom);
	document.removeEventListener('pointerup', onPointerUpFrom);
	document.removeEventListener('pointermove', onPointerMoveTo);
	document.removeEventListener('pointerup', onPointerUpTo);
});

watch(
	() => props.ui.value,
	(nextValue) => {
		valueFrom.value = nextValue?.from ?? 0;
		valueTo.value = nextValue?.to ?? 0;
	}
);
</script>

<style scoped>
.slider {
	@apply pt-8 pb-4 min-w-32;

	.bar {
		@apply relative w-full h-px;

		.bar-left {
			@apply absolute left-0 top-0 h-full bg-I0;
		}

		.bar-middle {
			@apply absolute top-0 h-full bg-I0;
		}

		.bar-right {
			@apply absolute right-0 top-0 h-full bg-current;
		}

		.step {
			@apply absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-px h-1 bg-current z-0;

			&.active {
				@apply bg-I0;
			}
		}

		.grabber {
			@apply absolute top-1/2 -translate-x-1/2 -translate-y-1/2 size-8 !rounded-full flex justify-center items-center cursor-grab z-10;

			.dot {
				content: '';
				@apply block size-3.5 rounded-full bg-I0;
			}

			.value {
				@apply absolute left-1/2 bottom-full -translate-x-1/2 text-I0 text-xs pb-px whitespace-nowrap;
			}

			&:focus,
			&:hover,
			&.dragging {
				@apply bg-I0 bg-opacity-20;
			}

			&.dragging {
				@apply cursor-grabbing;
			}
		}
	}

	&.range-mode {
		.bar {
			.bar-left {
				@apply bg-current;
			}
		}
	}

	&.error {
		.bar {
			.bar-left {
				@apply bg-SE0;
			}

			.bar-middle {
				@apply bg-SE0;
			}

			.grabber {
				.dot {
					@apply bg-SE0;
				}

				.value {
					@apply text-SE0;
				}

				&:focus,
				&:hover,
				&.dragging {
					@apply bg-SE0 bg-opacity-20;
				}
			}
		}
	}

	&.disabled {
		@apply text-SI0;

		.bar {
			.bar-left {
				@apply bg-ST0;
			}

			.bar-middle {
				@apply bg-ST0;
			}

			.step {
				@apply bg-SI0;

				&.active {
					@apply bg-ST0;
				}
			}

			.grabber {
				.dot {
					@apply bg-ST0;
				}

				.value {
					@apply text-ST0;
				}

				&:focus,
				&:hover,
				&.dragging {
					@apply bg-ST0 bg-opacity-20;
				}
			}
		}
	}
}
</style>
