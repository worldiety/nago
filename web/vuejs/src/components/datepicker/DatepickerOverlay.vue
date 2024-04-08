<template>
	<div v-if="expanded" ref="datepicker" class="fixed top-0 left-0 bottom-0 right-0 flex justify-center items-center text-black dark:text-white z-30">
		<div class="relative bg-white dark:bg-gray-700 rounded-xl shadow-lg max-w-96 p-6 z-10">
			<div class="h-[23rem]">
				<DatepickerHeader :label="label" @close="emit('submit', selectedStartDate, selectedEndDate)" class="mb-4" />

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
						class="relative flex justify-center items-center h-full w-full z-10"
						:class="{'within-range-day': datepickerDay.withinRange}"
					>
						<!-- TODO: Add click and keydown enter events to select start and end date -->
						<div
							class="day effect-hover relative flex justify-center items-center cursor-default z-10"
							:class="{
							'selected-day': datepickerDay.selected,
							'text-disabled-text': !datepickerDay.withinRange && datepickerDay.month !== currentMonthIndex + 1,
						}"
							tabindex="0"
						>
							<span>{{ datepickerDay.dayOfMonth }}</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Confirm button when in range mode -->
			<template v-if="rangeMode">
				<div class="border-b border-b-disabled-background mt-2 mb-4"></div>
				<button class="button-primary !text-black !w-full" @click="emit('submit', selectedStartDate, selectedEndDate)">{{ t('datepicker.confirm') }}</button>
			</template>
		</div>

		<!-- Blurred Background -->
		<div class="absolute top-0 left-0 bottom-0 right-0 backdrop-blur z-0" @click="emit('submit', selectedStartDate, selectedEndDate)"></div>
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
	(e: 'submit', selectedStartDate: Date, selectedEndDate: Date|null): void;
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
	const selectedStartDate = new Date();
	selectedStartDate.setFullYear(props.selectedStartYear, props.selectedStartMonth, props.selectedStartDay);
	selectedStartDate.setHours(0, 0, 0, 0);
	return selectedStartDate;
});

const selectedEndDate = computed((): Date => {
	const selectedEndDate = new Date();
	selectedEndDate.setFullYear(props.selectedEndYear, props.selectedEndMonth, props.selectedEndDay);
	selectedEndDate.setHours(0, 0, 0, 0);
	return selectedEndDate;
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

	const dayOfCurrentMonthDate = new Date();
	dayOfCurrentMonthDate.setFullYear(currentYear.value, currentMonthIndex.value + 1, 0);
	const lastDayOfCurrentMonth = dayOfCurrentMonthDate.getDate();
	for (let i = 1; i <= lastDayOfCurrentMonth; i++) {
		const dayOfWeekDate = new Date();
		dayOfWeekDate.setFullYear(currentYear.value, currentMonthIndex.value, i);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfWeekDate.getDay() === 0 ? 7 : dayOfWeekDate.getDay(),
			dayOfMonth: i,
			month: currentMonthIndex.value + 1,
			year: currentYear.value,
			selected: false,
			withinRange: false,
		};
		datepickerDay.selected = isSelectedDay(datepickerDay.dayOfMonth, datepickerDay.month, datepickerDay.year);
		datepickerDay.withinRange = isWithinRange(datepickerDay.dayOfMonth, datepickerDay.month, datepickerDay.year);
		daysOfCurrentMonth.push(datepickerDay);
	}

	return daysOfCurrentMonth;
}

function getFillingDaysOfPreviousMonth(): DatepickerDay[] {
	const fillingDaysOfPreviousMonth: DatepickerDay[] = [];

	const dayOfPreviousMonthDate = new Date();
	dayOfPreviousMonthDate.setFullYear(currentYear.value, currentMonthIndex.value, 0);
	const lastDayOfWeekPreviousMonth = dayOfPreviousMonthDate.getDay() === 0 ? 7 : dayOfPreviousMonthDate.getDay();
	const lastDayOfPreviousMonth = dayOfPreviousMonthDate.getDate();
	for (let i = 0; i < lastDayOfWeekPreviousMonth; i++) {
		dayOfPreviousMonthDate.setDate(lastDayOfPreviousMonth - i);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfPreviousMonthDate.getDay() === 0 ? 7 : dayOfPreviousMonthDate.getDay(),
			dayOfMonth: lastDayOfPreviousMonth - i,
			month: dayOfPreviousMonthDate.getMonth() + 1,
			year: dayOfPreviousMonthDate.getFullYear(),
			selected: false,
			withinRange: false,
		}
		datepickerDay.selected = isSelectedDay(datepickerDay.dayOfMonth, datepickerDay.month, datepickerDay.year);
		datepickerDay.withinRange = isWithinRange(datepickerDay.dayOfMonth, datepickerDay.month, datepickerDay.year);
		fillingDaysOfPreviousMonth.unshift(datepickerDay);
	}

	return fillingDaysOfPreviousMonth;
}

function getFillingDaysOfNextMonth(lastDayOfWeekCurrentMonth: number): DatepickerDay[] {
	const fillingDaysOfNextMonth: DatepickerDay[] = [];

	const dayOfNextMonthDate = new Date();
	if (currentMonthIndex.value + 1 === 12) {
		dayOfNextMonthDate.setFullYear(currentYear.value + 1, 0, 1);
	} else {
		dayOfNextMonthDate.setFullYear(currentYear.value, currentMonthIndex.value + 1, 1);
	}
	for (let i = 1; i <= 7 - lastDayOfWeekCurrentMonth; i++) {
		dayOfNextMonthDate.setDate(i);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfNextMonthDate.getDay() === 0 ? 7 : dayOfNextMonthDate.getDay(),
			dayOfMonth: dayOfNextMonthDate.getDate(),
			month: dayOfNextMonthDate.getMonth() + 1,
			year: dayOfNextMonthDate.getFullYear(),
			selected: false,
			withinRange: false,
		};
		datepickerDay.selected = isSelectedDay(datepickerDay.dayOfMonth, datepickerDay.month, datepickerDay.year);
		datepickerDay.withinRange = isWithinRange(datepickerDay.dayOfMonth, datepickerDay.month, datepickerDay.year);
		fillingDaysOfNextMonth.push(datepickerDay);
	}

	return fillingDaysOfNextMonth;
}

function isSelectedDay(day: number, month: number, year: number): boolean {
	return props.startDateSelected
		&& day === props.selectedStartDay
		&& month === props.selectedStartMonth
		&& year === props.selectedStartYear
		|| props.endDateSelected
		&& day === props.selectedEndDay
		&& month === props.selectedEndMonth
		&& year === props.selectedEndYear;
}

function isWithinRange(day: number, month: number, year: number): boolean {
	if (!props.startDateSelected || !props.endDateSelected || !selectedEndDate.value) {
		return false;
	}
	const dateToCheck = new Date();
	dateToCheck.setFullYear(year, month, day);
	dateToCheck.setHours(0, 0, 0, 0);
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
	@apply size-8 rounded-full
}

.selected-day {
	@apply bg-ora-orange bg-opacity-25 text-ora-orange;
}

.within-range-day {
	@apply bg-ora-orange bg-opacity-5 text-ora-orange;
}

/* Use no background when the first day of the selected range is within the last grid column */
.datepicker-grid > :not(.within-range-day) + div.within-range-day:nth-of-type(7n),
/* Use no background when the last day of the selected range is within the first grid column */
.datepicker-grid > .within-range-day:has(+ :not(.within-range-day)):nth-of-type(7n - 6) {
	@apply bg-white;
	@apply dark:bg-darkmode-gray;
}

/* Color the background of each day in the last grid column within the selected range */
.datepicker-grid > div.within-range-day:nth-of-type(7n) > .day::before {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-8 w-1/2 bg-ora-orange bg-opacity-5 rounded-r-full z-0;
}

/* Color the background of each day in the first grid column within the selected range */
.datepicker-grid > div.within-range-day:nth-of-type(7n - 6) > .day::before {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-8 w-1/2 bg-ora-orange bg-opacity-5 rounded-l-full z-0;
}

/* Round the last date within the selected range */
.datepicker-grid > .within-range-day:has(+ :not(.within-range-day)) {
	@apply relative rounded-r-full;
}

/* Round the first date within the selected range */
.datepicker-grid > :not(.within-range-day) + .within-range-day {
	@apply relative rounded-l-full;
}

/* Hide the background exceeding the right boundary of each day in the last column */
.datepicker-grid > div:nth-of-type(7n - 6)::before,
/* Hide the background exceeding the left boundary of the first day within the selected range */
.datepicker-grid > :not(.within-range-day) + .within-range-day::before {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-8 w-1/2 bg-white z-0;
	@apply dark:bg-darkmode-gray;
}

/* Hide the background exceeding the right boundary of each day in the last column */
.datepicker-grid > div:nth-of-type(7n)::before,
/* Hide the background exceeding the right boundary of the last day within the selected range */
.datepicker-grid > .within-range-day:has(+ :not(.within-range-day))::before {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-8 w-1/2 bg-white z-0;
	@apply dark:bg-darkmode-gray;
}
</style>
