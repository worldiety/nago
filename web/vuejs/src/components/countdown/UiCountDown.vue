<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed } from 'vue';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { CountDown, FunctionCallRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: CountDown;
}>();

const serviceAdapter = useServiceAdapter();

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

const duration = computed<Duration>(() => {
	let d = 0;
	if (props.ui.duration) {
		d = props.ui.duration;
	}
	return convertNanoseconds(d);
});

function invoke() {
	if (props.ui.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	}
}

function updateCountdown() {
	let d = 0;
	if (props.ui.duration) {
		d = props.ui.duration;
	}

	d -= 1;
	if (d <= 0) {
		d = 0;
	}

	props.ui.duration = d;
}

let intervalId: number | undefined = 0;
intervalId = setInterval(() => {
	if (props.ui.duration !== undefined) {
		if (props.ui.duration > 0) {
			updateCountdown();
			//console.log("countdown update");
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

function formatWithLeadingZero(value: number): string {
	return value.toString().padStart(2, '0');
}
</script>

<template v-if="props.ui.iv">
	<div class="text-center flex flex-col md:flex-row md:space-x-8" :style="frameStyles">
		<div class="flex justify-center space-x-8 grow">
			<div v-if="props.ui.showDays" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(duration.days) }}</p>
				<p class="text-lg">Tage</p>
			</div>
			<div class="border-l mt-2 mb-2" :style="borderStyle"></div>
			<div v-if="props.ui.showHours" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(duration.hours) }}</p>
				<p class="text-lg">Stunden</p>
			</div>
		</div>
		<div class="border-l hidden md:block mt-2 mb-2" :style="borderStyle"></div>
		<div class="flex justify-center space-x-8 mt-4 md:mt-0 grow">
			<div v-if="props.ui.showMinutes" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(duration.minutes) }}</p>
				<p class="text-lg">Minuten</p>
			</div>
			<div class="border-l mt-2 mb-2" :style="borderStyle"></div>
			<div v-if="props.ui.showSeconds" class="grow">
				<p class="text-6xl font-bold">{{ formatWithLeadingZero(duration.seconds) }}</p>
				<p class="text-lg">Sekunden</p>
			</div>
		</div>
	</div>
</template>
