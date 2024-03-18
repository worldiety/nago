<script setup lang="ts">
import type { LiveDatepicker } from '@/shared/model/liveDatepicker';
import { computed, ref } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import ArrowDown from '@/assets/svg/arrowDown.svg';
import { useNetworkStore } from '@/stores/networkStore';
import monthNames from '@/shared/monthNames';

const props = defineProps<{
	ui: LiveDatepicker;
}>();

const networkStore = useNetworkStore();
const currentDate = new Date(Date.now());
const currentDay = ref<number>(currentDate.getDate());
const currentMonthIndex = ref<number>(currentDate.getMonth());
const currentYear = ref<number>(currentDate.getFullYear());

const totalDaysInMonth = computed((): number => {
	const lastDayOfMonthDate = new Date();
	lastDayOfMonthDate.setFullYear(currentYear.value, currentMonthIndex.value + 1, 0);
	return lastDayOfMonthDate.getDate();
});

const dayStartOffsetInMonth = computed((): number => {
	const firstDayOfMonthDate = new Date();
	firstDayOfMonthDate.setFullYear(currentYear.value, currentMonthIndex.value, 1);
	return firstDayOfMonthDate.getDay() - 1;
});

const dateFormatted = computed((): string => {
	const date = new Date();
	date.setFullYear(currentYear.value, currentMonthIndex.value, currentDay.value);
	return date.toLocaleDateString();
});

function isInCurrentMonth(day: number): boolean {
	return day == currentDay.value && currentMonthIndex.value == currentDate.getMonth() && currentYear.value == currentDate.getFullYear();
}

function decreaseMonth(): void {
	if (currentMonthIndex.value === 0) {
		currentMonthIndex.value = 11;
		currentYear.value -= 1;
		return;
	}
	currentMonthIndex.value -= 1;
}

function increaseMonth(): void {
	if (currentMonthIndex.value === 11) {
		currentMonthIndex.value = 0;
		currentYear.value += 1;
		return;
	}
	currentMonthIndex.value += 1;
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</span>
		<div class="relative">
			<!-- Input field -->
			<div class="relative z-0">
				<input
					:value="dateFormatted"
					type="text"
					readonly
					:disabled="props.ui.disabled.value"
					class="input-field w-full pr-8"
					@click="networkStore.invokeFunc(props.ui.onToggleExpanded)"
				>
				<div class="absolute top-0 bottom-0 right-2 flex items-center pointer-events-none h-full">
					<Calendar class="w-4" :class="props.ui.disabled.value ? 'text-disabled-text' : 'text-black'" />
				</div>
			</div>

			<!-- Datepicker -->
			<div v-if="props.ui.expanded.value" class="absolute top-8 right-8 bg-white rounded-md shadow-lg p-2 z-10">
				<div class="flex justify-between items-center mb-4">
					<div class="effect-hover flex justify-center items-center rounded-full size-8" @click="decreaseMonth">
						<ArrowDown class="rotate-90 h-4" />
					</div>
					<p class="text-center">{{ monthNames.get(currentMonthIndex) }} {{ currentYear }}</p>
					<div class="effect-hover flex justify-center items-center rounded-full size-8" @click="increaseMonth">
						<ArrowDown class="-rotate-90 h-4" />
					</div>
				</div>

				<div class="grid grid-cols-7 gap-2 text-center leading-none">
					<span>Mo</span>
					<span>Di</span>
					<span>Mi</span>
					<span>Do</span>
					<span>Fr</span>
					<span>Sa</span>
					<span>So</span>

					<div v-for="(_offset, index) in dayStartOffsetInMonth" :key="index"></div>
					<div v-for="(day, index) in totalDaysInMonth" :key="index" class="flex justify-center items-center h-full w-full">
						<div class="day effect-hover flex justify-center items-center cursor-default" :class="{'selected-day': isInCurrentMonth(day)}">
							<span>{{ day }}</span>
						</div>
					</div>
				</div>
			</div>
		</div>
		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>

<style scoped>
.day {
	@apply size-8 rounded-full
}

.selected-day {
	@apply text-wdy-green;
}
</style>
