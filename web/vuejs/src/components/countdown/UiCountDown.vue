<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { CountDown, FunctionCallRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: CountDown;
}>();

const serviceAdapter = useServiceAdapter();

const progress = ref<number>(0);
const duration = ref(props.ui.duration);
const initialDuration = ref<number | undefined>(undefined);

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);
	if (props.ui.textColor) {
		styles.push(`color: ${colorValue(props.ui.textColor)}`);
	}
	return styles.join(';');
});

interface Duration {
	days: number;
	hours: number;
	minutes: number;
	seconds: number;
}

function convertNanoseconds(seconds: number): Duration {
	//console.log(seconds);
	const minutes = Math.floor(seconds / 60);
	const hours = Math.floor(minutes / 60);
	const days = Math.floor(hours / 24);

	const remainingSeconds = seconds % 60;
	const remainingMinutes = minutes % 60;
	const remainingHours = hours % 24;

	return {
		days: days,
		hours: remainingHours,
		minutes: remainingMinutes,
		seconds: remainingSeconds,
	};
}

const borderStyle = computed<string>(() => {
	if (props.ui.separatorColor) {
		return `border-color: ${colorValue(props.ui.separatorColor)}`;
	}

	return '';
});

const durationNano = computed<Duration>(() => {
	return convertNanoseconds(duration.value ?? 0);
});

//
const progressBackgroundStyle = computed<string>(() => {
	if (!props.ui.progressBackground) {
		return `background-color: var(--M6)`;
	}

	return `background-color: ${colorValue(props.ui.progressBackground)}`;
});

// note: do not make it a computed property, because vue will ignore progress value changes
function progressForegroundStyle(): string {
	if (!props.ui.progressColor) {
		return `background-color: var(--I0); width: ${progress.value}%`;
	}

	return `background-color: ${colorValue(props.ui.progressColor)}; width: ${progress.value}%`;
}

function invoke() {
	if (props.ui.action && !props.ui.done) {
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	}
}

function updateCountdown() {
	if (!initialDuration.value) return;

	let d = 0;
	if (duration.value) {
		d = duration.value;
	}

	d -= 1;
	if (d <= 0) {
		d = 0;
	}

	duration.value = d;
	progress.value = (duration.value / initialDuration.value) * 100;
}

let intervalId: any | undefined = 0;

function formatWithLeadingZero(value: number): string {
	return value.toString().padStart(2, '0');
}

onMounted(() => {
	intervalId = setInterval(() => {
		if (duration.value !== undefined) {
			if (duration.value > 0) {
				updateCountdown();
			} else {
				invoke();
				console.log('countdown stopped');
				if (intervalId !== undefined) {
					clearInterval(intervalId);
				}
			}
		}
		//console.log("tick");
	}, 1000);

	initialDuration.value = duration.value;
	progress.value = 100;

	if (props.ui.done) {
		duration.value = 0;
		progress.value = 0;
		return;
	}
});

onUnmounted(() => {
	clearInterval(intervalId);
});

watch(
	() => props.ui.duration,
	() => {
		duration.value = props.ui.duration;
	}
);
</script>

<template v-if="props.ui.iv">
	<div v-if="!props.ui.style" class="text-center flex flex-col md:flex-row md:space-x-8" :style="frameStyles">
		<div class="flex justify-center space-x-8 grow">
			<div v-if="props.ui.showDays" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(durationNano.days) }}</p>
				<p class="text-lg">Tage</p>
			</div>
			<div class="border-l mt-2 mb-2" :style="borderStyle"></div>
			<div v-if="props.ui.showHours" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(durationNano.hours) }}</p>
				<p class="text-lg">Stunden</p>
			</div>
		</div>
		<div class="border-l hidden md:block mt-2 mb-2" :style="borderStyle"></div>
		<div class="flex justify-center space-x-8 mt-4 md:mt-0 grow">
			<div v-if="props.ui.showMinutes" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(durationNano.minutes) }}</p>
				<p class="text-lg">Minuten</p>
			</div>
			<div class="border-l mt-2 mb-2" :style="borderStyle"></div>
			<div v-if="props.ui.showSeconds" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(durationNano.seconds) }}</p>
				<p class="text-lg">Sekunden</p>
			</div>
		</div>
	</div>

	<!-- Countdown Progress bar -->
	<div v-else class="flex-1 space-y-1 overflow-hidden" :style="frameStyles">
		<div class="w-full h-1.5 rounded-full overflow-hidden mt-1" :style="progressBackgroundStyle">
			<div class="h-full transition-all duration-300" :style="progressForegroundStyle()"></div>
		</div>
	</div>
</template>
