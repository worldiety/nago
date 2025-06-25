<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div
		v-if="expanded"
		ref="datepicker"
		class="fixed top-0 left-0 bottom-0 right-0 flex justify-center items-center z-30"
		@keydown.tab.exact="moveFocusForward"
		@keydown.shift.tab="moveFocusBackwards"
	>
		<div class="relative bg-M1 rounded-xl shadow-lg max-w-96 p-6 z-10">
			<div class="h-[23rem]">
				<DatepickerHeader ref="datepickerHeader" :label="label" @close="emit('close')" class="mb-4" />

				<!-- Datepicker content -->
				<div class="flex justify-between items-center mb-4 h-8">
					<div
						class="hover:bg-I0/15 flex justify-center items-center cursor-pointer rounded-full size-8"
						tabindex="0"
						@click="decreaseMonth"
						@keydown.enter="decreaseMonth"
					>
						<ArrowRight class="rotate-180 h-4" />
					</div>
					<div class="flex justify-center items-center basis-2/3 gap-x-px text-lg h-full">
						<div class="basis-1/2 shrink-0 grow-0 h-full">
							<select
								v-model="currentMonthIndex"
								class="hover:bg-I0/15 border-0 bg-M1 text-right cursor-pointer rounded-l-md w-full h-full px-2"
							>
								<option
									v-for="(monthEntry, index) of monthNames.entries()"
									:key="index"
									:value="monthEntry[0]"
								>
									{{ monthEntry[1] }}
								</option>
							</select>
						</div>
						<div class="basis-1/2 shrink-0 grow-0 h-full">
							<input
								v-model="yearInput"
								type="text"
								class="hover:bg-I0/15 border-0 bg-M1 rounded-r-md text-left w-full h-full px-2"
							/>
						</div>
					</div>
					<div
						class="hover:bg-I0/15 flex justify-center items-center cursor-pointer rounded-full size-8"
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
							:ref="(el) => setLastDatepickerDayElement(el, index)"
							class="day hover:bg-I0/15 flex justify-center items-center cursor-pointer"
							:class="{
								'selected-day': datepickerDay.selectedStart || datepickerDay.selectedEnd,
								'text-disabled-text':
									!datepickerDay.withinRange && datepickerDay.monthIndex !== currentMonthIndex,
							}"
							tabindex="0"
							@click="selectDate(datepickerDay)"
							@keydown.enter="selectDate(datepickerDay)"
						>
							<span class="select-none">{{ datepickerDay.dayOfMonth }}</span>
						</div>
					</div>
				</div>
			</div>

			<template v-if="rangeMode">
				<!-- Hint texts and clear button when in range mode -->
				<p v-if="rangeSelectionState === RangeSelectionState.SELECT_START" class="mt-2">
					Bitte wählen Sie einen Startzeitpunkt aus
				</p>
				<p v-else-if="rangeSelectionState === RangeSelectionState.SELECT_END" class="mt-2">
					Bitte wählen Sie einen Endzeitpunkt aus
				</p>
				<button
					v-else-if="rangeSelectionState === RangeSelectionState.COMPLETE"
					@click="$emit('clearSelection')"
					class="flex justify-start items-center gap-x-2 text-I0 underline mt-2"
				>
					<undo-icon class="h-4" /> Auswahl aufheben
				</button>

				<div class="border-b border-b-disabled-background mt-3 mb-6"></div>

				<!-- Confirm button when in range mode -->
				<button
					ref="confirmButton"
					class="button-confirm button-primary"
					:disabled="rangeSelectionState !== RangeSelectionState.COMPLETE"
					@click="emit('submitSelection')"
				>
					{{ t('datepicker.confirm') }}
				</button>
			</template>
		</div>

		<!-- Blurred Background -->
		<div class="absolute top-0 left-0 bottom-0 right-0 bg-opacity-60 bg-black z-0" @click="emit('close')"></div>
	</div>
</template>

<script setup lang="ts">
import { ComponentPublicInstance, computed, nextTick, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import ArrowRight from '@/assets/svg/arrowRightBold.svg';
import UndoIcon from '@/assets/svg/undo.svg';
import DatepickerHeader from '@/components/datepicker/DatepickerHeader.vue';
import type DatepickerDay from '@/components/datepicker/datepickerDay';
import { RangeSelectionState } from '@/components/datepicker/rangeSelectionState';
import monthNames from '@/shared/monthNames';

const props = defineProps<{
	expanded: boolean;
	rangeMode: boolean;
	label?: string;
	selectedStartDay: number;
	selectedStartMonth: number;
	selectedStartYear: number;
	selectedEndDay: number;
	selectedEndMonth: number;
	selectedEndYear: number;
	rangeSelectionState: RangeSelectionState;
}>();

const emit = defineEmits<{
	(e: 'close'): void;
	(e: 'select', day: number, month: number, year: number): void;
	(e: 'submitSelection'): void;
	(e: 'clearSelection'): void;
}>();

const { t } = useI18n();
const datepickerHeader = useTemplateRef('datepickerHeader');
const confirmButton = useTemplateRef('confirmButton');
const datepicker = ref<HTMLElement | undefined>();
const currentDate = new Date(Date.now());
const currentYear = ref<number>(currentDate.getFullYear());
const currentMonthIndex = ref<number>(currentDate.getMonth());
const yearInput = ref<string>(currentYear.value.toString(10));
const lastDatepickerDayIndex = ref<number | null>(null);
const lastDatepickerDayElement = ref<ComponentPublicInstance | Element | null>(null);

watch(
	() => props.expanded,
	(newValue) => {
		if (newValue) {
			nextTick(() => datepickerHeader.value?.closeButton?.focus());
		}
	}
);

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
	return new Date(props.selectedStartYear, props.selectedStartMonth - 1, props.selectedStartDay, 0, 0, 0, 0);
});

const selectedEndDate = computed((): Date => {
	return new Date(props.selectedEndYear, props.selectedEndMonth - 1, props.selectedEndDay, 0, 0, 0, 0);
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

function setLastDatepickerDayElement(datepickerDay: ComponentPublicInstance | Element | null, index: number) {
	// Update the last element if its index is greater than the index of the current last element or the index of the
	// current last element is null
	if (lastDatepickerDayIndex.value === null || index > lastDatepickerDayIndex.value) {
		lastDatepickerDayIndex.value = index;
		lastDatepickerDayElement.value = datepickerDay;
	}
}

function getDaysOfCurrentMonth(): DatepickerDay[] {
	const daysOfCurrentMonth: DatepickerDay[] = [];

	const dayOfCurrentMonthDate = new Date(currentYear.value, currentMonthIndex.value + 1, 0, 0, 0, 0, 0);
	const lastDayOfCurrentMonth = dayOfCurrentMonthDate.getDate();
	for (let i = 1; i <= lastDayOfCurrentMonth; i++) {
		const dayOfWeekDate = new Date(currentYear.value, currentMonthIndex.value, i, 0, 0, 0, 0);
		const datepickerDay: DatepickerDay = {
			dayOfWeek: dayOfWeekDate.getDay() === 0 ? 7 : dayOfWeekDate.getDay(),
			dayOfMonth: i,
			monthIndex: currentMonthIndex.value,
			year: currentYear.value,
			selectedStart: false,
			selectedEnd: false,
			withinRange: false,
		};
		datepickerDay.selectedStart = isSelectedStartDay(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		datepickerDay.selectedEnd = isSelectedEndDay(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		datepickerDay.withinRange = isWithinRange(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		daysOfCurrentMonth.push(datepickerDay);
	}

	return daysOfCurrentMonth;
}

function getFillingDaysOfPreviousMonth(): DatepickerDay[] {
	const fillingDaysOfPreviousMonth: DatepickerDay[] = [];

	const dayOfPreviousMonthDate = new Date(currentYear.value, currentMonthIndex.value, 0, 0, 0, 0, 0);
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
		};
		datepickerDay.selectedStart = isSelectedStartDay(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		datepickerDay.selectedEnd = isSelectedEndDay(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		datepickerDay.withinRange = isWithinRange(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		fillingDaysOfPreviousMonth.unshift(datepickerDay);
	}

	return fillingDaysOfPreviousMonth;
}

function getFillingDaysOfNextMonth(lastDayOfWeekCurrentMonth: number): DatepickerDay[] {
	const fillingDaysOfNextMonth: DatepickerDay[] = [];

	let dayOfNextMonthDate: Date;
	if (currentMonthIndex.value + 1 === 12) {
		dayOfNextMonthDate = new Date(currentYear.value + 1, 0, 1, 0, 0, 0, 0);
	} else {
		dayOfNextMonthDate = new Date(currentYear.value, currentMonthIndex.value + 1, 1, 0, 0, 0, 0);
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
		datepickerDay.selectedStart = isSelectedStartDay(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		datepickerDay.selectedEnd = isSelectedEndDay(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		datepickerDay.withinRange = isWithinRange(
			datepickerDay.dayOfMonth,
			datepickerDay.monthIndex,
			datepickerDay.year
		);
		fillingDaysOfNextMonth.push(datepickerDay);
	}

	return fillingDaysOfNextMonth;
}

function isSelectedStartDay(day: number, monthIndex: number, year: number): boolean {
	return (
		props.rangeSelectionState > RangeSelectionState.SELECT_START &&
		day === props.selectedStartDay &&
		monthIndex === props.selectedStartMonth - 1 &&
		year === props.selectedStartYear
	);
}

function isSelectedEndDay(day: number, monthIndex: number, year: number): boolean {
	return (
		props.rangeSelectionState > RangeSelectionState.SELECT_END &&
		day === props.selectedEndDay &&
		monthIndex === props.selectedEndMonth - 1 &&
		year === props.selectedEndYear
	);
}

function isWithinRange(day: number, monthIndex: number, year: number): boolean {
	if (props.rangeSelectionState !== RangeSelectionState.COMPLETE) {
		return false;
	}

	//console.log("check range",day,monthIndex,year)
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

function selectDate(datepickerDay: DatepickerDay): void {
	emit('select', datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
	if (!props.rangeMode) {
		emit('close');
	}
}

function moveFocusForward(event: Event) {
	if (document.activeElement === getLastFocusableElement()) {
		event.preventDefault();
		datepickerHeader.value?.closeButton?.focus();
	}
}

function getLastFocusableElement(): HTMLElement | null {
	if (props.rangeSelectionState === RangeSelectionState.COMPLETE) {
		return confirmButton.value;
	}
	if (lastDatepickerDayElement.value instanceof HTMLElement) {
		return lastDatepickerDayElement.value;
	}
	return null;
}

function moveFocusBackwards(event: Event) {
	if (document.activeElement === datepickerHeader.value?.closeButton) {
		event.preventDefault();
		getLastFocusableElement()?.focus();
	}
}
</script>

<style scoped>
.day {
	@apply relative size-8 rounded-full z-20;
}

.selected-day {
	@apply bg-I0 bg-opacity-25 text-M8;
}

.within-range-day {
	@apply bg-I0 bg-opacity-5 text-M8;
}

/* Each day in the first column within the selection range except the selected days (after element) */
.datepicker-grid
	> .within-range-day:nth-of-type(7n - 6):not(.selected-start-day-container, .selected-start-day-container)::after {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-full w-1/2 bg-M1;
}

/* Each day in the first grid column within the selected range that is not a selected day (before element) */
.datepicker-grid > .within-range-day:nth-of-type(7n - 6) > .day:not(.selected-day)::before {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-full w-1/2 bg-I0 bg-opacity-5 rounded-l-full;
}

/* Each day in the last column within the selected range except the selected days (before element) */
.datepicker-grid
	> .within-range-day:nth-of-type(7n):not(.selected-start-day-container, .selected-end-day-container)::before {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-full w-1/2 bg-M1;
}

/* Each day in the last grid column within the selected range that is not a selected day (after element) */
.datepicker-grid > .within-range-day:nth-of-type(7n) > .day:not(.selected-day)::after {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-full w-1/2 bg-I0 bg-opacity-5 rounded-r-full;
}

/* First day of selected range (before element) */
.datepicker-grid > .selected-start-day-container::before {
	content: '';
	width: calc(50% + 1rem);
	@apply absolute top-0 left-0 bottom-0 bg-M1 rounded-r-full h-full;
}

/* Last day of selected range (after element) */
.datepicker-grid > .selected-end-day-container::after {
	content: '';
	width: calc(50% + 1rem);
	@apply absolute top-0 bottom-0 right-0 bg-M1 rounded-l-full h-full;
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
</style>
