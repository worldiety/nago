<template>
	<div v-if="expanded" ref="datepicker" class="fixed top-0 left-0 bottom-0 right-0 flex justify-center items-center text-black dark:text-white z-30">
		<div class="relative bg-white dark:bg-gray-700 rounded-xl shadow-lg max-w-96 p-6 z-10">
			<div class="h-[23rem]">
				<DatepickerHeader :label="label" @close="emit('close')" class="mb-4" />

				<!-- Datepicker content -->
				<div class="flex justify-between items-center mb-4 h-8">
					<div
						class="effect-hover flex justify-center items-center rounded-full size-8"
						tabindex="0"
						@click="decreaseMonth"
						@keydown.enter="decreaseMonth"
					>
						<ArrowRight class="rotate-180 h-4" />
					</div>
					<div class="flex justify-center items-center basis-2/3 gap-x-px text-lg h-full">
						<div class="basis-1/2 shrink-0 grow-0 h-full">
							<select v-model="currentMonthIndex" class="effect-hover border-0 bg-white dark:bg-darkmode-gray text-right cursor-default rounded-l-md w-full h-full px-2">
								<option v-for="(monthEntry, index) of monthNames.entries()" :key="index" :value="monthEntry[0]">
									{{ monthEntry[1] }}
								</option>
							</select>
						</div>
						<div class="basis-1/2 shrink-0 grow-0 h-full">
							<input v-model="yearInput" type="text" class="effect-hover border-0 bg-white dark:bg-darkmode-gray rounded-r-md text-left w-full h-full px-2">
						</div>
					</div>
					<div
						class="effect-hover flex justify-center items-center rounded-full size-8"
						tabindex="0"
						@click="increaseMonth"
						@keydown.enter="increaseMonth"
					>
						<ArrowRight class="h-4" />
					</div>
				</div>

				<div class="datepicker-grid grid grid-cols-7 gap-y-2 text-center leading-none">
					<span>Mo</span>
					<span>Di</span>
					<span>Mi</span>
					<span>Do</span>
					<span>Fr</span>
					<span>Sa</span>
					<span>So</span>

					<div
						v-for="(datepickerDay, index) in datepickerDays"
						:key="index"
						class="relative flex justify-center items-center h-full w-full"
						:class="{
							'within-range-day': datepickerDay.withinRange,
							'selected-start-day-container': datepickerDay.selectedStart,
							'selected-end-day-container': datepickerDay.selectedEnd,
						}"
					>
						<div
							class="day effect-hover flex justify-center items-center cursor-default"
							:class="{
							'selected-day': datepickerDay.selectedStart || datepickerDay.selectedEnd,
							'text-disabled-text': !datepickerDay.withinRange && datepickerDay.monthIndex !== currentMonthIndex,
						}"
							tabindex="0"
							@click="emit('select', datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year)"
							@keydown.enter="emit('select', datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year)"
						>
							<span>{{ datepickerDay.dayOfMonth }}</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Confirm button when in range mode -->
			<template v-if="rangeMode">
				<div class="border-b border-b-disabled-background mt-2 mb-4"></div>
				<button
					class="button-confirm button-primary"
					:disabled="!startDateSelected || !endDateSelected"
					@click="emit('close')"
				>
					{{ t('datepicker.confirm') }}
				</button>
			</template>
		</div>

		<!-- Blurred Background -->
		<div class="absolute top-0 left-0 bottom-0 right-0 backdrop-blur z-0" @click="emit('close')"></div>
	</div>
</template>

<script setup lang="ts">
import monthNames from '@/shared/monthNames'
import ArrowRight from '@/assets/svg/arrowRightBold.svg';
import { computed, ref, watch } from 'vue';
import DatepickerHeader from '@/components/datepicker/DatepickerHeader.vue';
import type DatepickerDay from '@/components/datepicker/datepickerDay';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
	expanded: boolean;
	rangeMode: boolean;
	label: string;
	startDateSelected: boolean;
	selectedStartDay: number;
	selectedStartMonth: number;
	selectedStartYear: number;
	endDateSelected: boolean;
	selectedEndDay: number;
	selectedEndMonth: number;
	selectedEndYear: number;
}>();

const emit = defineEmits<{
	(e: 'close'): void;
	(e: 'select', day: number, month: number, year: number): void;
}>();

const { t } = useI18n();
const datepicker = ref<HTMLElement|undefined>();
const currentDate = new Date(Date.now());
const currentYear = ref<number>(currentDate.getFullYear());
const currentMonthIndex = ref<number>(currentDate.getMonth());
const yearInput = ref<string>(currentYear.value.toString(10));

/**
 * Only allow year values with a length between 1 and 4.
 * Does also prevent values less than 1 and greater than 9999.
 */
watch(yearInput, (newValue, oldValue) => {
	if (newValue.match(/^[1-9][0-9]{0,3}$/)) {
		currentYear.value = parseInt(newValue, 10);
	} else {
		yearInput.value = oldValue;
	}
});

const selectedStartDate = computed((): Date => {
	return new Date(
		props.selectedStartYear,
		props.selectedStartMonth - 1,
		props.selectedStartDay,
		0,
		0,
		0,
		0,
	);
});

const selectedEndDate = computed((): Date => {
	return new Date(
		props.selectedEndYear,
		props.selectedEndMonth - 1,
		props.selectedEndDay,
		0,
		0,
		0,
		0,
	);
});

const datepickerDays = computed((): DatepickerDay[] => {
	const datepickerDays: DatepickerDay[] = [];

	// Add days of current month
	datepickerDays.push(...getDaysOfCurrentMonth());

	// Add filling days of previous month, if the current month's first day is not a monday
	if (datepickerDays[0].dayOfWeek !== 1) {
		datepickerDays.unshift(...getFillingDaysOfPreviousMonth());
	}

	// Add filling days of next month, if the current month's last day is not a sunday
	const lastDayOfWeekCurrentMonth = datepickerDays.at(-1)?.dayOfWeek;
	if (lastDayOfWeekCurrentMonth !== undefined && lastDayOfWeekCurrentMonth !== 7) {
		datepickerDays.push(...getFillingDaysOfNextMonth(lastDayOfWeekCurrentMonth));
	}

	return datepickerDays;
});

function getDaysOfCurrentMonth(): DatepickerDay[] {
	const daysOfCurrentMonth: DatepickerDay[] = [];

	const dayOfCurrentMonthDate = new Date(
		currentYear.value,
		currentMonthIndex.value + 1,
		0,
		0,
		0,
		0,
		0,
	);
	const lastDayOfCurrentMonth = dayOfCurrentMonthDate.getDate();
	for (let i = 1; i <= lastDayOfCurrentMonth; i++) {
		const dayOfWeekDate = new Date(
			currentYear.value,
			currentMonthIndex.value,
			i,
			0,
			0,
			0,
			0,
		);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfWeekDate.getDay() === 0 ? 7 : dayOfWeekDate.getDay(),
			dayOfMonth: i,
			monthIndex: currentMonthIndex.value,
			year: currentYear.value,
			selectedStart: false,
			selectedEnd: false,
			withinRange: false,
		};
		datepickerDay.selectedStart = isSelectedStartDay(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		datepickerDay.selectedEnd = isSelectedEndDay(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		datepickerDay.withinRange = isWithinRange(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		daysOfCurrentMonth.push(datepickerDay);
	}

	return daysOfCurrentMonth;
}

function getFillingDaysOfPreviousMonth(): DatepickerDay[] {
	const fillingDaysOfPreviousMonth: DatepickerDay[] = [];

	const dayOfPreviousMonthDate = new Date(
		currentYear.value,
		currentMonthIndex.value,
		0,
		0,
		0,
		0,
		0,
	);
	const lastDayOfWeekPreviousMonth = dayOfPreviousMonthDate.getDay() === 0 ? 7 : dayOfPreviousMonthDate.getDay();
	const lastDayOfPreviousMonth = dayOfPreviousMonthDate.getDate();
	for (let i = 0; i < lastDayOfWeekPreviousMonth; i++) {
		dayOfPreviousMonthDate.setDate(lastDayOfPreviousMonth - i);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfPreviousMonthDate.getDay() === 0 ? 7 : dayOfPreviousMonthDate.getDay(),
			dayOfMonth: lastDayOfPreviousMonth - i,
			monthIndex: dayOfPreviousMonthDate.getMonth(),
			year: dayOfPreviousMonthDate.getFullYear(),
			selectedStart: false,
			selectedEnd: false,
			withinRange: false,
		}
		datepickerDay.selectedStart = isSelectedStartDay(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		datepickerDay.selectedEnd = isSelectedEndDay(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		datepickerDay.withinRange = isWithinRange(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		fillingDaysOfPreviousMonth.unshift(datepickerDay);
	}

	return fillingDaysOfPreviousMonth;
}

function getFillingDaysOfNextMonth(lastDayOfWeekCurrentMonth: number): DatepickerDay[] {
	const fillingDaysOfNextMonth: DatepickerDay[] = [];

	let dayOfNextMonthDate: Date;
	if (currentMonthIndex.value + 1 === 12) {
		dayOfNextMonthDate = new Date(
			currentYear.value + 1,
			0,
			1,
			0,
			0,
			0,
			0,
		);
	} else {
		dayOfNextMonthDate = new Date(
			currentYear.value,
			currentMonthIndex.value + 1,
			1,
			0,
			0,
			0,
			0,
		);
	}
	for (let i = 1; i <= 7 - lastDayOfWeekCurrentMonth; i++) {
		dayOfNextMonthDate.setDate(i);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfNextMonthDate.getDay() === 0 ? 7 : dayOfNextMonthDate.getDay(),
			dayOfMonth: dayOfNextMonthDate.getDate(),
			monthIndex: dayOfNextMonthDate.getMonth(),
			year: dayOfNextMonthDate.getFullYear(),
			selectedStart: false,
			selectedEnd: false,
			withinRange: false,
		};
		datepickerDay.selectedStart = isSelectedStartDay(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		datepickerDay.selectedEnd = isSelectedEndDay(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		datepickerDay.withinRange = isWithinRange(datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
		fillingDaysOfNextMonth.push(datepickerDay);
	}

	return fillingDaysOfNextMonth;
}

function isSelectedStartDay(day: number, monthIndex: number, year: number): boolean {
	return props.startDateSelected
		&& day === props.selectedStartDay
		&& monthIndex === props.selectedStartMonth - 1
		&& year === props.selectedStartYear;
}

function isSelectedEndDay(day: number, monthIndex: number, year: number): boolean {
	return props.endDateSelected
	&& day === props.selectedEndDay
	&& monthIndex === props.selectedEndMonth - 1
	&& year === props.selectedEndYear;
}

function isWithinRange(day: number, monthIndex: number, year: number): boolean {
	if (!props.startDateSelected || !props.endDateSelected) {
		return false;
	}
	const dateToCheck = new Date(year, monthIndex, day, 0, 0, 0, 0);
	return selectedStartDate.value <= dateToCheck && dateToCheck <= selectedEndDate.value;
}

function decreaseMonth(): void {
	if (currentMonthIndex.value === 0) {
		currentMonthIndex.value = 11;
		currentYear.value -= 1;
		yearInput.value = currentYear.value.toString(10);
		return;
	}
	currentMonthIndex.value -= 1;
}

function increaseMonth(): void {
	if (currentMonthIndex.value === 11) {
		currentMonthIndex.value = 0;
		currentYear.value += 1;
		yearInput.value = currentYear.value.toString(10);
		return;
	}
	currentMonthIndex.value += 1;
}
</script>

<style scoped>
.day {
	@apply relative size-8 rounded-full z-20;
}

.selected-day {
	@apply bg-ora-orange bg-opacity-25 text-ora-orange;
}

.within-range-day {
	@apply bg-ora-orange bg-opacity-5 text-ora-orange;
}

/* Each day in the first column within the selection range except the selected days (after element) */
.datepicker-grid > .within-range-day:nth-of-type(7n - 6):not(.selected-start-day-container, .selected-start-day-container)::after {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-full w-1/2 bg-white;
	@apply dark:bg-darkmode-gray;
}

/* Each day in the first grid column within the selected range that is not a selected day (before element) */
.datepicker-grid > .within-range-day:nth-of-type(7n - 6) > .day:not(.selected-day)::before {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-full w-1/2 bg-ora-orange bg-opacity-5 rounded-l-full;
}

/* Each day in the last column within the selected range except the selected days (before element) */
.datepicker-grid > .within-range-day:nth-of-type(7n):not(.selected-start-day-container, .selected-end-day-container)::before {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-full w-1/2 bg-white;
	@apply dark:bg-darkmode-gray;
}

/* Each day in the last grid column within the selected range that is not a selected day (after element) */
.datepicker-grid > .within-range-day:nth-of-type(7n) > .day:not(.selected-day)::after {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-full w-1/2 bg-ora-orange bg-opacity-5 rounded-r-full;
}

/* First day of selected range (before element) */
.datepicker-grid > .selected-start-day-container::before {
	content: '';
	width: calc(50% + 1rem);
	@apply absolute top-0 left-0 bottom-0 bg-white rounded-r-full h-full;
	@apply dark:bg-darkmode-gray;
}

/* Last day of selected range (after element) */
.datepicker-grid > .selected-end-day-container::after {
	content: '';
	width: calc(50% + 1rem);
	@apply absolute top-0 bottom-0 right-0 bg-white rounded-l-full h-full;
	@apply dark:bg-darkmode-gray;
}

/* Selected start day container in last grid row */
.datepicker-grid > .within-range-day:nth-of-type(7n).selected-start-day-container,
/* Selected end day container in first grid row */
.datepicker-grid > .within-range-day:nth-of-type(7n - 6).selected-end-day-container {
	@apply bg-transparent;
}

.button-confirm {
	@apply w-full;
}

.button-confirm:not(:disabled) {
	@apply text-black;
}
</style>
